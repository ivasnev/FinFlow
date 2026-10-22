import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from parser import parse_receipt, money


def lines(*texts):
    return [{'text': text, 'confidence': .99} for text in texts]


def test_products_quantities_and_total():
    result = parse_receipt(lines('Молоко 89,99', 'Хлеб', '2 x 35,00 = 70,00', 'ИТОГО 159,99', 'НАЛИЧНЫМИ 200,00', 'СДАЧА 40,01'))
    assert [(item['name'], item['amount_cents']) for item in result['items']] == [('Молоко', 8999), ('Хлеб', 7000)]
    assert result['total_cents'] == 15999
    assert result['items'][1]['quantity'] == '2'
    assert not result['warnings']


def test_money_precision_and_boundaries():
    assert money('1 234,03') == 123403
    assert money('0.01') == 1
    assert money('99999999999999.99') is None
    assert money('invalid') is None


def test_mismatch_and_uncertain_lines_require_review():
    result = parse_receipt([{'text': 'Молоко 89,99', 'confidence': .7}, {'text': 'ИТОГО 100,00', 'confidence': .99}])
    assert len(result['warnings']) == 2
    assert result['items'][0]['amount_cents'] == 8999


def test_no_invented_items_from_totals_taxes_and_payment():
    result = parse_receipt(lines('ИНН 1234567890', 'НДС 10,00', 'ИТОГО 100,00', 'БЕЗНАЛИЧНЫМИ 100,00', 'Карта 1234'))
    assert not result['items']
    assert result['total_cents'] == 10000


def test_bounded_result_for_untrusted_ocr():
    result = parse_receipt(lines(*['Товар 1,00'] * 2000))
    assert len(result['items']) == 100
    assert len(result['raw_text']) <= 100000


def test_unit_quantity_is_not_mistaken_for_total_price():
    result = parse_receipt(lines('Хлеб 129.99 *3 =383.C7 HDC 10%', 'Лимоны 169.99 *0.936 =159.11 HDC 10%', 'ИТОГО 549.08'))
    assert [item['amount_cents'] for item in result['items']] == [38997, 15911]
    assert result['items'][0]['quantity'] == '3'
    assert result['items'][0]['confidence'] < .85
    assert result['warnings']


def test_bank_metadata_and_discount_columns_are_not_goods():
    result = parse_receipt(lines('Молоко', 'HAC 10% 259.99 30.00', '229.99*1 шт 229.99', 'Шоколад', 'НДС 20% 104.99 104.99*1', 'WT 104.99', 'Итог с учетом скидок:', '=334.98', 'НТОГ 30.00', 'Cyммa(Pyб) 334.98'))
    assert [(i['name'], i['amount_cents']) for i in result['items']] == [('Молоко', 22999), ('Шоколад', 10499)]
    assert result['total_cents'] == 33498


def test_total_on_next_line_and_embedded_bank_slip():
    result = parse_receipt(lines('Сушки 49.99', 'Cyммa(Pyб) 109.98', 'Сухари 59.99', 'ИТОГО', '109.98', 'СУММА НДС 10% 10.00'))
    assert len(result['items']) == 2
    assert result['total_cents'] == 10998


def test_per_item_discount_changes_item_but_not_grand_total():
    result = parse_receipt(lines('Покупка первая 120.00', 'Итог с учетом скидок: 100.00', 'Покупка вторая 50.00', 'ИТОГО 150.00'))
    assert [i['amount_cents'] for i in result['items']] == [10000, 5000]
    assert result['total_cents'] == 15000


def test_ambiguous_thousands_are_not_silently_rewritten_to_match_items():
    result = parse_receipt(lines("Покупка 334.98", "ИТОГО 4 334.98"))
    assert result["total_cents"] == 433498
    assert result["warnings"]


def test_separate_name_keeps_its_low_confidence_and_source():
    result = parse_receipt([{'text':'Молоко', 'confidence':.2}, {'text':'89.99', 'confidence':.99}, {'text':'ИТОГО 89.99', 'confidence':.99}])
    assert result['items'][0]['confidence'] == .2
    assert result['items'][0]['source_text'] == 'Молоко\n89.99'
    assert result['warnings']


def test_english_service_lines_are_filtered_before_transliteration():
    result = parse_receipt(lines('SUBTOTAL 100.00', 'CASH 100.00', 'TAX 20.00', 'TOTAL 100.00'))
    assert result['items'] == []
    assert result['total_cents'] == 10000


def test_discount_inference_preserves_source_and_warns_even_if_sum_matches():
    result = parse_receipt(lines('Первый товар 200.00', 'Второй товар 400.00', 'Итог с учетом скидок: 300.00', 'ИТОГО 500.00'))
    assert result['items'][1]['amount_cents'] == 30000
    assert result['items'][1]['confidence'] < .85
    assert 'Итог с учетом скидок' in result['items'][1]['source_text']
    assert result['warnings']


def test_per_item_discount_equal_to_price_does_not_end_the_receipt():
    result = parse_receipt(lines('Первый товар 100.00', 'Итог с учетом скидок: 100.00', 'Второй товар 50.00', 'ИТОГО 150.00'))
    assert len(result['items']) == 2
    assert result['total_cents'] == 15000


def test_utf8_result_stays_within_backend_byte_limits():
    result = parse_receipt(lines(*['Я'*1000]*1200))
    assert len(result['raw_text'].encode()) <= 100000
