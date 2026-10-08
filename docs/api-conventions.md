# Params of get list 

    page  int  
    limit int // max is 100

## without it will be return 
    page = 1
    limit = 20

| Ситуация | Статус | Код ошибки |
|---|---:|---|
| Неправильный JSON | 400 | `invalid_json` |
| Неправильные поля или единица | 400 | `validation_error` |
| Неправильная пагинация | 400 | `invalid_pagination` |
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