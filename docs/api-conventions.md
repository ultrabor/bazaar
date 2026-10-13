# Params of get list 

    page  int  
    limit int // max is 100

## Product filters

`GET /product` принимает необязательные query-параметры:

- `name` — регистронезависимый поиск буквальной подстроки. Символы `%`, `_` и `\` не считаются шаблонами.
- `sku` — точное совпадение SKU.
- `unit` — точное совпадение: `kg`, `meter`, `piece` или `liter`.

Несколько фильтров объединяются через `AND` и работают вместе с `page` и `limit`.
Если совпадений нет, API возвращает `200` и `[]`.

## without it will be return 
    page = 1
    limit = 20

| Ситуация | Статус | Код ошибки |
|---|---:|---|
| Неправильный JSON | 400 | `invalid_json` |
| Неправильные поля или единица | 400 | `validation_error` |
| Неправильная пагинация | 400 | `invalid_pagination` |
| Неправильный формат UUID товара | 400 | `invalid_product_id` |
| Товар недоступен или отсутствует | 404 | `product_not_found` |
| Дубликат SKU | 409 | `sku_conflict` |
| Неожиданная внутренняя ошибка | 500 | `internal_error` |
| Таймаут зависимости | 503 | `service_unavailable` |

# Error format
```json
{
  "error": {
    "code": "invalid_pagination",
    "message": "limit must be between 1 and 100"
  }
}
```
