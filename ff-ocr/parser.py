"""Разбор кассовых строк без автоматического создания финансовых операций."""
import re
from decimal import Decimal, InvalidOperation, ROUND_HALF_UP

MONEY = re.compile(r'(?<![\d.,])([\d]+(?:[ \u00a0]\d{3})*[.,]\s?\d{2})(?!\d)')
FOOTER = re.compile(r'^(?:итог|итого|всего|к оплате|total|grand total|amount due)\b', re.I)
SERVICE = re.compile(r'^(?:инн|фн\b|фд\b|фп\b|ндс|нас\b|налог|сдача|наличн|безнал|банковск|кассир|кассовый|терминал|приход|оплачено|получено|subtotal|sub total|tax|change|cash|visa|mastercard|payment|rounding|скидк|discount)', re.I)
QUANTITY = re.compile(r'^\s*(\d+(?:[.,]\d{1,3})?)\s*[xх*×]\s*', re.I)
PRICE_QUANTITY = re.compile(r'(\d+[.,]\s?\d{2})\s*[xх*×]\s*(\d+(?:[.,]\d{1,3})?)(?!\d)', re.I)
VAT = re.compile(r'\b(?:НДС|НАС|HAC|HДC|HDC|HC)\s*\d+(?:[.,]\d+)?\s*%', re.I)


def money(text):
    try:
        value = int((Decimal(text.replace(' ', '').replace('\u00a0', '').replace(',', '.')) * 100).quantize(Decimal(1), rounding=ROUND_HALF_UP))
        return value if 0 <= value <= 9_999_999_999 else None
    except (InvalidOperation, ValueError):
        return None


def normalized(text):
    # Только классификация реквизитов: написание названий товаров сохраняется.
    return text.upper().translate(str.maketrans('ABCEHKMOPTXY', 'АВСЕНКМОРТХУ'))


def parse_receipt(lines):
    items, total, pending, warnings = [], None, '', []
    pending_confidence, pending_source = 1.0, ''
    discount_inference = False
    finished, awaiting_total, awaiting_discount = False, False, False
    for line in lines[:1200]:
        text = str(line.get('text', '')).strip()[:1000]
        confidence = max(0.0, min(1.0, float(line.get('confidence', 0))))
        source_text = text
        normal = normalized(text)
        compact = re.sub(r'[^А-ЯЁA-Z]', '', normal)
        matches = list(MONEY.finditer(text))
        # Первый явный итог завершает товарную часть. Банковский слип ниже не является покупками.
        is_discount_total = compact.startswith('ИТОГСУЧЕТОМСКИДОК')
        is_total = compact.startswith(('ИТОГО', 'ИТОГ', 'КОПЛАТЕ')) or bool(FOOTER.match(text))
        fuzzy_total = bool(items) and compact.startswith(('НТОГ', 'ТОГОКОПЛАТЕ'))
        if fuzzy_total and total is not None:
            finished, pending = True, ''
            continue
        is_total = is_total or fuzzy_total
        if is_discount_total or awaiting_discount:
            if matches and items:
                value = money(matches[-1].group(1))
                if value is not None and value > 0:
                    total = value
                    discount_inference = True
                    if value < items[-1]['amount_cents']:
                        items[-1]['amount_cents'] = value
                        items[-1]['confidence'] = min(items[-1]['confidence'], confidence, .8)
                        items[-1]['source_text'] = (items[-1]['source_text'] + '\n' + source_text).encode()[:4000].decode('utf-8', errors='ignore')
                awaiting_discount = False
            else:
                awaiting_discount = True
            pending = ''
            continue
        if is_total and not finished:
            total = money(matches[-1].group(1)) if matches else None
            finished, awaiting_total, pending = True, not bool(matches), ''
            continue
        if awaiting_total and matches:
            total = money(matches[-1].group(1))
            awaiting_total = False
        if finished:
            continue
        # Ставка НДС бывает отдельной строкой ПЕРЕД ценой, либо суффиксом товарной строки.
        vat = VAT.search(text)
        vat_prefix = vat is not None and not text[:vat.start()].strip()
        if vat:
            text = (text[vat.end():] if vat_prefix else text[:vat.start()]).strip()
            matches = list(MONEY.finditer(text))
            if vat_prefix and not PRICE_QUANTITY.search(text) and '=' not in text:
                # Табличные «цена до скидки / скидка» не подменяют конечную цену следующей строки.
                continue
        normal = normalized(text)
        if SERVICE.match(text) or SERVICE.match(normal) or re.match(r'^(?:ДАТА|ВРЕМЯ|АДРЕС|САЙТ|WWW\.|HTTP|КАРТА|СУММА|СYММА|БАЛАНС|БАПАНС|ВАША|БАША|НАЧИСЛЕНО|НАЧИСЛЕН|НАКОПЛЕНО|ОКРУГЛЕНИЕ|ОДОБРЕНО|RRN|KIA|MIR|МЕСТО РАСЧЕТОВ|ККТ|АВТОМАТ|АКЦИОНЕРНОЕ|СПАСИБО)', normal):
            pending = ''
            continue
        if not matches:
            if any(c.isalpha() for c in text) and len(text) >= 3 and not re.search(r'\d{8,}', text):
                pending, pending_confidence, pending_source = text, confidence, source_text
            continue
        value = money(matches[-1].group(1))
        prefix = text[:matches[0].start()].strip(' .:=*-')
        prefix = re.sub(r'^\d+[.):]\s*', '', prefix)
        qty_match = QUANTITY.match(text)
        inline_name = any(c.isalpha() for c in prefix) and not qty_match
        name = prefix if inline_name else pending
        if not inline_name and pending:
            confidence = min(confidence, pending_confidence)
            source_text = pending_source + '\n' + source_text
        quantity = qty_match.group(1).replace(',', '.') if qty_match else None
        price_qty = PRICE_QUANTITY.search(text)
        if price_qty:
            quantity = price_qty.group(2).replace(',', '.')
            # Если печатная сумма не прочитана, используем прочитанные цену и количество,
            # обязательно понижая уверенность: скидка на позицию могла потеряться.
            after = text[price_qty.end():]
            amount_match = list(MONEY.finditer(after))
            if not amount_match:
                unit = money(price_qty.group(1))
                value = int((Decimal(unit) * Decimal(quantity)).quantize(Decimal(1), rounding=ROUND_HALF_UP)) if unit is not None else None
                confidence = min(confidence, .8)
        if not name or sum(c.isalpha() for c in name) < 3 or value is None or not 0 < value <= 9_999_999_999:
            continue
        if re.search(r'^[\d\s.,хx*×=]+$', name):
            continue
        items.append({'name': name[:255], 'amount_cents': value, 'confidence': confidence,
                      'quantity': quantity, 'source_text': source_text.encode()[:4000].decode('utf-8', errors='ignore')})
        pending = ''
        if len(items) >= 100:
            warnings.append('Распознаны первые 100 позиций. Проверьте полноту чека.')
            break
    if discount_inference:
        warnings.append('Строка итога со скидками может относиться к позиции или всему чеку. Проверьте суммы по фото.')
    if not items:
        warnings.append('Не удалось уверенно выделить покупки. Добавьте позиции вручную или загрузите более чёткое фото.')
    if total is None:
        warnings.append('Итог чека не найден. Сверьте сумму с фотографией.')
    elif sum(item['amount_cents'] for item in items) != total:
        warnings.append('Сумма позиций не совпадает с итогом чека. Проверьте скидки, весовые товары и пропущенные строки.')
    if any(item['confidence'] < 0.85 for item in items):
        warnings.append('Есть строки с низкой уверенностью распознавания или расчётом по цене и количеству. Сверьте их с фото.')
    return {'items': items, 'total_cents': total, 'warnings': warnings,
            'raw_text': '\n'.join(str(line.get('text', ''))[:1000] for line in lines[:1200])[:100_000].encode()[:100_000].decode('utf-8', errors='ignore')}
