"""Submit the dependencies actually built/tested, with local Go replacements respected.

Collection is read-only and needs no token. Publishing is restricted to verified
artifacts of the same main commit; failed/partial scans must leave the old graph intact.
"""
import argparse
import datetime
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import urllib.request
from urllib.parse import quote

ROOT = Path(__file__).resolve().parents[2]
DETECTOR = 'finflow-resolved-dependencies'
REQUIREMENTS = ('ff-ocr/requirements.txt', 'ff-ocr/requirements-test.txt')


def run(*args, cwd=ROOT):
    return subprocess.check_output(args, cwd=cwd, text=True)


def json_stream(text):
    decoder = json.JSONDecoder()
    while text.strip():
        text = text.lstrip()
        value, end = decoder.raw_decode(text)
        yield value
        text = text[end:]


def purl(ecosystem, name, version):
    return f'pkg:{ecosystem}/{quote(name, safe="/")}@{quote(version, safe="")}'


def go_entry(package, scope):
    if package.get('Error') or package.get('DepsErrors'):
        raise ValueError('Go could not resolve ' + package.get('ImportPath', 'package'))
    module = package.get('Module')
    if not module or module.get('Main'):
        return None
    actual = module.get('Replace', module)
    if 'Replace' in module and not actual.get('Version'):
        directory = Path(actual['Dir']).resolve()
        if not directory.is_relative_to(ROOT):
            raise ValueError('A local Go replacement is outside the checked-out repository')
        # The source is already in this checkout, not the historical remote pseudo-version.
        # Its third-party imports are still collected through go list -deps.
        return None
    if not actual.get('Version'):
        raise ValueError('A third-party Go module has no resolved version')
    url = purl('golang', actual['Path'], actual['Version'])
    return url, {'package_url': url, 'scope': scope,
                 'relationship': 'indirect' if module.get('Indirect') else 'direct'}


def collect_go(module):
    directory = ROOT / module
    if module not in {p.parent.name for p in ROOT.glob('ff-*/go.mod')}:
        raise ValueError('Unknown Go module')
    resolved = {}

    def include(patterns, scope, tests=False):
        args = ['go', 'list', '-mod=readonly', '-deps', '-json']
        if tests:
            args.append('-test')
        for package in json_stream(run(*args, *patterns, cwd=directory)):
            entry = go_entry(package, scope)
            if entry:
                key, value = entry
                if key not in resolved or scope == 'runtime':
                    resolved[key] = value

    include(['./cmd/...'] if (directory / 'cmd').exists() else ['./...'], 'runtime')
    include(['./...'], 'development', tests=True)
    tools = run('go', 'list', 'tool', cwd=directory).split()
    if tools:
        include(tools, 'development')
    if not resolved:
        raise ValueError('Unexpectedly empty Go dependency graph')
    manifest = module + '/go.mod'
    return {manifest: {'name': manifest, 'file': {'source_location': manifest},
                       'resolved': dict(sorted(resolved.items()))}}


def pip_closure(roots, report):
    from packaging.requirements import Requirement
    from packaging.utils import canonicalize_name
    if report.get('version') != '1':
        raise ValueError('Unsupported pip inspect report')
    installed = {canonicalize_name(p['metadata']['name']): p for p in report['installed']}
    environment = report['environment']
    extras, visited, edges, direct = {}, set(), {}, set()
    pending = []

    def require(req, parent=None):
        name = canonicalize_name(req.name)
        if req.url:
            raise ValueError('URL dependencies need an explicit source audit')
        if name not in installed:
            raise ValueError('Missing Python dependency: ' + name)
        item = installed[name]
        if item.get('direct_url'):
            raise ValueError('Python dependency was installed outside the package index: ' + name)
        if not req.specifier.contains(item['metadata']['version'], prereleases=True):
            raise ValueError('Installed version does not satisfy requirement: ' + str(req))
        if parent is None:
            direct.add(name)
        else:
            edges.setdefault(parent, set()).add(name)
        old = extras.setdefault(name, set())
        new = set(req.extras) - old
        old.update(req.extras)
        if name not in visited or new:
            pending.append(name)

    for line in roots:
        req = Requirement(line)
        if not req.marker or req.marker.evaluate({**environment, 'extra': ''}):
            require(req)
    while pending:
        name = pending.pop()
        visited.add(name)
        for value in installed[name]['metadata'].get('requires_dist', []):
            req = Requirement(value)
            if not req.marker or any(req.marker.evaluate({**environment, 'extra': extra})
                                     for extra in extras[name] | {''}):
                require(req, name)
    urls = {name: purl('pypi', name, installed[name]['metadata']['version']) for name in visited}
    return {urls[name]: {'package_url': urls[name],
                        'relationship': 'direct' if name in direct else 'indirect',
                        'dependencies': sorted(urls[child] for child in edges.get(name, ()))}
            for name in sorted(visited)}


def collect_pip():
    report = json.loads(run(sys.executable, '-m', 'pip', 'inspect', '--local'))
    manifests = {}
    for filename in REQUIREMENTS:
        roots = [line.split('#', 1)[0].strip() for line in (ROOT / filename).read_text().splitlines()]
        resolved = pip_closure([line for line in roots if line], report)
        if not resolved:
            raise ValueError('Unexpectedly empty Python dependency graph')
        for dependency in resolved.values():
            dependency['scope'] = 'development' if filename.endswith('-test.txt') else 'runtime'
        manifests[filename] = {'name': filename, 'file': {'source_location': filename}, 'resolved': resolved}
    return manifests, report['environment']


def snapshot(manifests, correlator, metadata=None):
    sha = run('git', 'rev-parse', 'HEAD').strip()
    if os.environ.get('GITHUB_SHA', sha) != sha:
        raise ValueError('Checkout does not match the workflow commit')
    metadata = {**(metadata or {}), 'source_clean': not bool(run('git', 'status', '--porcelain', '--untracked-files=all').strip())}
    return {'version': 0, 'sha': sha, 'ref': os.environ.get('GITHUB_REF', 'refs/heads/main'),
            'job': {'id': os.environ.get('GITHUB_RUN_ID', sha), 'correlator': correlator},
            'detector': {'name': DETECTOR, 'version': '1.0.0',
                         'url': 'https://github.com/ivasnev/FinFlow/tree/main/scripts/security'},
            'scanned': datetime.datetime.now(datetime.timezone.utc).isoformat(), 'manifests': manifests,
            'metadata': metadata or {}}


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


def publish(directory):
    if os.environ.get('GITHUB_REF') != 'refs/heads/main' or os.environ.get('GITHUB_EVENT_NAME') not in {'push', 'schedule', 'workflow_dispatch'}:
        raise ValueError('Only a trusted main workflow may publish dependencies')
    repository = os.environ['GITHUB_REPOSITORY']
    if not re.fullmatch(r'[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+', repository):
        raise ValueError('Invalid repository')
    sha = run('git', 'rev-parse', 'HEAD').strip()
    if sha != os.environ['GITHUB_SHA']:
        raise ValueError('Checkout does not match the workflow commit')
    expected = {'finflow-go-' + p.parent.name: {str(p.relative_to(ROOT))}
                for p in ROOT.glob('ff-*/go.mod')}
    for machine in ('x86_64', 'aarch64'):
        expected['finflow-pip-ocr-linux-' + machine] = set(REQUIREMENTS)
    snapshots = [json.loads(p.read_text()) for p in sorted(directory.glob('*.json'))]
    correlators = set()
    for data in snapshots:
        if data['sha'] != sha or data['ref'] != 'refs/heads/main' or data['detector']['name'] != DETECTOR or data.get('metadata', {}).get('source_clean') is not True:
            raise ValueError('Snapshot belongs to another commit or detector')
        correlator = data['job']['correlator']
        if correlator in correlators or set(data['manifests']) != expected.get(correlator):
            raise ValueError('Missing, unexpected or duplicate dependency snapshot')
        if correlator.startswith('finflow-pip-ocr-') and data['metadata'].get('platform') != correlator.removeprefix('finflow-pip-ocr-'):
            raise ValueError('Python snapshot platform does not match its collector')
        for name, manifest in data['manifests'].items():
            if manifest['file']['source_location'] != name or not manifest['resolved']:
                raise ValueError('Invalid or empty manifest')
        correlators.add(correlator)
    if correlators != set(expected):
        raise ValueError('Refusing to publish an incomplete dependency graph')
    opener = urllib.request.build_opener(NoRedirect())
    for data in snapshots:
        # Publish jobs are serialized by the workflow; a slow old build must not
        # overwrite newer main snapshots after finishing out of order.
        reference = urllib.request.Request(f'https://api.github.com/repos/{repository}/git/ref/heads/main',
            headers={'Authorization': 'Bearer ' + os.environ['GITHUB_TOKEN'],
                     'Accept': 'application/vnd.github+json'})
        with opener.open(reference, timeout=30) as response:
            if json.load(response)['object']['sha'] != sha:
                print('Skipping dependency publication: main has advanced.', flush=True)
                return
        request = urllib.request.Request(f'https://api.github.com/repos/{repository}/dependency-graph/snapshots',
            data=json.dumps(data).encode(), headers={'Authorization': 'Bearer ' + os.environ['GITHUB_TOKEN'],
            'Accept': 'application/vnd.github+json', 'Content-Type': 'application/json',
            'X-GitHub-Api-Version': '2022-11-28'}, method='POST')
        with opener.open(request, timeout=30) as response:
            if response.status != 201:
                raise ValueError('GitHub did not accept the dependency snapshot')
            result = json.load(response)
        print(data['job']['correlator'] + ': submitted snapshot ' + str(result['id']), flush=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('mode', choices=['go', 'pip', 'publish'])
    parser.add_argument('--module')
    parser.add_argument('--output', type=Path)
    parser.add_argument('--directory', type=Path)
    args = parser.parse_args()
    if args.mode == 'publish':
        publish(args.directory)
        return
    if args.mode == 'go':
        manifests = collect_go(args.module)
        correlator, metadata = 'finflow-go-' + args.module, {'coverage': 'application-tests-tools'}
    else:
        manifests, environment = collect_pip()
        target = environment['platform_system'].lower() + '-' + environment['platform_machine'].lower()
        correlator, metadata = 'finflow-pip-ocr-' + target, {'platform': target, 'python': environment['python_version']}
    args.output.write_text(json.dumps(snapshot(manifests, correlator, metadata), indent=2) + '\n')
    print('Collected ' + str(sum(len(m['resolved']) for m in manifests.values())) + ' resolved dependencies')


if __name__ == '__main__':
    main()
