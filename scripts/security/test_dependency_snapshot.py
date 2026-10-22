import io
import json
import os
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

from dependency_snapshot import ROOT, collect_go, go_entry, pip_closure, publish


def package(name, version, dependencies=(), **extra):
    return {'metadata': {'name': name, 'version': version, 'requires_dist': list(dependencies)}, **extra}


class DependencySnapshotTests(unittest.TestCase):
    def test_remote_replacement_uses_the_actual_selected_source(self):
        module = {'Path': 'example.org/old', 'Version': 'v1.0.0', 'Indirect': True,
                  'Replace': {'Path': 'example.org/fixed', 'Version': 'v2.0.0+patch'}}
        key, result = go_entry({'Module': module}, 'development')
        self.assertEqual(key, 'pkg:golang/example.org/fixed@v2.0.0%2Bpatch')
        self.assertEqual(result['scope'], 'development')
        self.assertNotIn('old', key)

    def test_local_replacement_is_not_a_stale_remote_release(self):
        module = {'Path': 'example.org/old', 'Version': 'v0.0.0-old',
                  'Replace': {'Path': '../ff-id', 'Dir': str(ROOT / 'ff-id')}}
        self.assertIsNone(go_entry({'Module': module}, 'runtime'))
        module['Replace']['Dir'] = '/private/tmp/external-source'
        with self.assertRaises(ValueError):
            go_entry({'Module': module}, 'runtime')

    def test_go_test_and_explicit_tool_dependencies_remain_in_the_graph(self):
        def output(*args, **kwargs):
            if args == ('go', 'list', 'tool'):
                return 'example.org/generator/cmd/tool\n'
            name = 'application' if './cmd/...' in args else 'test-helper' if '-test' in args else 'generator'
            return '{"Module":{"Path":"example.org/' + name + '","Version":"v2.0.0"}}'
        with patch('dependency_snapshot.run', side_effect=output):
            resolved = collect_go('ff-auth')['ff-auth/go.mod']['resolved']
        self.assertEqual(len(resolved), 3)
        self.assertEqual(resolved['pkg:golang/example.org/application@v2.0.0']['scope'], 'runtime')
        self.assertEqual(resolved['pkg:golang/example.org/generator@v2.0.0']['scope'], 'development')
        self.assertEqual(resolved['pkg:golang/example.org/test-helper@v2.0.0']['scope'], 'development')

    def report(self):
        return {'version': '1', 'environment': {'platform_system': 'Linux', 'extra': ''}, 'installed': [
            package('root', '1.0', ['shared>=2', 'extra-dep; extra == "ocr"', 'windows-only; platform_system == "Windows"']),
            package('shared', '2.1'), package('extra-dep', '3.0', ['shared>=2']),
            package('unrelated-audit-tool', '0.1'),
        ]}

    def test_pip_includes_extras_and_transitive_dependencies_but_not_runner_tools(self):
        resolved = pip_closure(['root[ocr]==1.0'], self.report())
        self.assertEqual(set(resolved), {'pkg:pypi/root@1.0', 'pkg:pypi/shared@2.1', 'pkg:pypi/extra-dep@3.0'})
        self.assertEqual(resolved['pkg:pypi/root@1.0']['relationship'], 'direct')
        self.assertEqual(resolved['pkg:pypi/shared@2.1']['relationship'], 'indirect')
        self.assertIn('pkg:pypi/extra-dep@3.0', resolved['pkg:pypi/root@1.0']['dependencies'])

    def test_pip_expands_an_extra_requested_later_through_another_dependency(self):
        report = self.report()
        report['installed'].append(package('other', '1.0', ['root[ocr]']))
        self.assertIn('pkg:pypi/extra-dep@3.0', pip_closure(['root', 'other'], report))

    def test_pip_resolution_failure_is_not_published_as_an_empty_or_partial_success(self):
        with self.assertRaises(ValueError):
            pip_closure(['root==2.0'], self.report())
        report = self.report()
        report['installed'][0]['metadata']['requires_dist'].append('missing')
        with self.assertRaises(ValueError):
            pip_closure(['root'], report)
        report = self.report()
        report['installed'][1]['direct_url'] = {'url': 'https://unverified.invalid/source.zip'}
        with self.assertRaises(ValueError):
            pip_closure(['root'], report)

    def test_publish_refuses_pull_requests_before_reading_the_token_or_network(self):
        with patch.dict(os.environ, {'GITHUB_REF': 'refs/pull/1/merge', 'GITHUB_EVENT_NAME': 'pull_request'}, clear=True):
            with self.assertRaises(ValueError):
                publish(Path('/does-not-exist'))

    def test_publish_refuses_missing_scan_artifacts(self):
        environment = {'GITHUB_REF': 'refs/heads/main', 'GITHUB_EVENT_NAME': 'push',
                       'GITHUB_REPOSITORY': 'ivasnev/FinFlow', 'GITHUB_SHA': 'a' * 40}
        with tempfile.TemporaryDirectory() as directory, patch.dict(os.environ, environment, clear=True):
            with patch('dependency_snapshot.run', return_value='a' * 40):
                with self.assertRaisesRegex(ValueError, 'incomplete'):
                    publish(Path(directory))

    def test_slow_old_main_build_cannot_overwrite_a_newer_graph(self):
        environment = {'GITHUB_REF': 'refs/heads/main', 'GITHUB_EVENT_NAME': 'push',
                       'GITHUB_REPOSITORY': 'ivasnev/FinFlow', 'GITHUB_SHA': 'a' * 40,
                       'GITHUB_TOKEN': 'synthetic-token'}
        calls = []
        class Opener:
            def open(self, request, **kwargs):
                calls.append(request.get_method())
                if request.get_method() != 'GET':
                    raise AssertionError('A stale main build must never submit a snapshot')
                return io.BytesIO(json.dumps({'object': {'sha': 'b' * 40}}).encode())
        with tempfile.TemporaryDirectory() as directory, patch.dict(os.environ, environment, clear=True):
            for module in ('ff-auth', 'ff-id', 'ff-split', 'ff-files', 'ff-common', 'ff-tvm'):
                self.write_snapshot(directory, 'finflow-go-' + module, [module + '/go.mod'])
            for machine in ('x86_64', 'aarch64'):
                self.write_snapshot(directory, 'finflow-pip-ocr-linux-' + machine,
                                    ['ff-ocr/requirements.txt', 'ff-ocr/requirements-test.txt'],
                                    {'platform': 'linux-' + machine})
            with patch('dependency_snapshot.run', return_value='a' * 40), patch('dependency_snapshot.urllib.request.build_opener', return_value=Opener()):
                publish(Path(directory))
        self.assertEqual(calls, ['GET'])

    def test_publisher_rejects_missing_arm_coverage(self):
        environment = {'GITHUB_REF': 'refs/heads/main', 'GITHUB_EVENT_NAME': 'push',
                       'GITHUB_REPOSITORY': 'ivasnev/FinFlow', 'GITHUB_SHA': 'a' * 40}
        with tempfile.TemporaryDirectory() as directory, patch.dict(os.environ, environment, clear=True):
            for module in ('ff-auth', 'ff-id', 'ff-split', 'ff-files', 'ff-common', 'ff-tvm'):
                self.write_snapshot(directory, 'finflow-go-' + module, [module + '/go.mod'])
            self.write_snapshot(directory, 'finflow-pip-ocr-linux-x86_64',
                                ['ff-ocr/requirements.txt', 'ff-ocr/requirements-test.txt'],
                                {'platform': 'linux-x86_64'})
            with patch('dependency_snapshot.run', return_value='a' * 40):
                with self.assertRaisesRegex(ValueError, 'incomplete'):
                    publish(Path(directory))

    def test_publisher_rejects_dirty_local_source_marked_as_a_commit_snapshot(self):
        environment = {'GITHUB_REF': 'refs/heads/main', 'GITHUB_EVENT_NAME': 'push',
                       'GITHUB_REPOSITORY': 'ivasnev/FinFlow', 'GITHUB_SHA': 'a' * 40}
        with tempfile.TemporaryDirectory() as directory, patch.dict(os.environ, environment, clear=True):
            self.write_snapshot(directory, 'finflow-go-ff-auth', ['ff-auth/go.mod'], {'source_clean': False})
            with patch('dependency_snapshot.run', return_value='a' * 40):
                with self.assertRaisesRegex(ValueError, 'another commit'):
                    publish(Path(directory))

    def write_snapshot(self, directory, correlator, files, metadata=None):
        manifests = {name: {'file': {'source_location': name}, 'resolved': {'test': {'package_url': 'pkg:pypi/test@1.0'}}} for name in files}
        data = {'sha': 'a' * 40, 'ref': 'refs/heads/main', 'detector': {'name': 'finflow-resolved-dependencies'},
                'job': {'correlator': correlator}, 'manifests': manifests, 'metadata': {'source_clean': True, **(metadata or {})}}
        (Path(directory) / (correlator + '.json')).write_text(json.dumps(data))


if __name__ == '__main__':
    unittest.main()
