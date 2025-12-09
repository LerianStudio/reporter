# Reporter Template Technical Reference

> **Last Updated**: November 23, 2025
> **Source**: Lerian Reporter Documentation

---

## Template Model

Reporter templates use `.tpl` extension and mirror the exact structure of the final output:
- XML-structured template → XML output
- HTML-structured template → HTML/PDF output
- TXT-structured template → TXT output
- CSV-structured template → CSV output

**Important**: File content follows output format but MUST use `.tpl` extension.

## Placeholders Syntax

Templates use placeholders to fetch data dynamically:

```django
{{ base.table_or_collection.field_or_document }}
```

Works for both SQL databases (tables) and MongoDB (collections).

## Template Blocks

### Loop
```django
{% for item in list %}
  ...
{% endfor %}
```

### Conditional
```django
{% if value_a == value_b %}
  ...
{% endif %}
```

### Temporary Scope
```django
{% with object as alias %}
  ...
{% endwith %}
```

### Value Formatting
```django
{{ field_name|floatformat:2 }}  # Renders as 123.45
```

## Conditional Operations

| Operator | Description | Example |
|----------|-------------|---------|
| `==` | Equal | `{% if a == b %}` |
| `!=` | Not equal | `{% if a != b %}` |
| `>` | Greater than | `{% if a > b %}` |
| `<` | Less than | `{% if a < b %}` |
| `>=` | Greater/equal | `{% if a >= b %}` |
| `<=` | Less/equal | `{% if a <= b %}` |
| `and` | Both true | `{% if a and b %}` |
| `or` | At least one true | `{% if a or b %}` |
| `not` | Inverts Boolean | `{% if not a %}` |

## Math Functions

| Function | Description | Example |
|----------|-------------|---------|
| `sum_by` | Sum values | `{% sum_by transaction.operation by "amount" if condition %}` |
| `count_by` | Count items | `{% count_by transaction.operation if condition %}` |
| `avg_by` | Calculate average | `{% avg_by transaction.operation by "amount" if condition %}` |
| `min_by` | Minimum value | `{% min_by transaction.operation by "amount" if condition %}` |
| `max_by` | Maximum value | `{% max_by transaction.operation by "amount" if condition %}` |
| `percent_of` | Calculate percentage | `{{ category.amount \| percent_of: total.expenses }}` |
| `filter()` | Filter lists | `filter(list, "field", value)` |
| `{% calc %}` | Inline calculations | `{% calc (balance.available + 1.2) * balance.on_hold %}` |

## Date and Time

Current date/time when rendering:
```django
{% date_time "dd/MM/YYYY HH:mm" %}
```
Generated in UTC without regional adjustments.

## String Filters

| Filter | Description | Example |
|--------|-------------|---------|
| `slice` | Extract substring | `{{ cnpj \| slice:":8" }}` |
| `upper` | Convert to uppercase | `{{ text \| upper }}` |
| `lower` | Convert to lowercase | `{{ text \| lower }}` |
| `ljust` | Left align with padding | `{{ text \| ljust:"10" }}` |
| `rjust` | Right align with padding | `{{ text \| rjust:"10" }}` |
| `center` | Center with padding | `{{ text \| center:"10" }}` |
| `floatformat` | Format decimal places | `{{ value \| floatformat:2 }}` |
| `date` | Format date | `{{ date \| date:"Y-m-d" }}` |

## Contains Function

Check partial inclusion:
```django
{% if contains(midaz_transaction.transaction.body.source.from.account_alias, midaz_onboarding.account.alias) %}
```

## Template Examples

### XML Template - CADOC 4010

```xml
<?xml version="1.0" encoding="UTF-8"?>
<documento codigoDocumento="4010" cnpj="{{ organization.legal_document|slice:':8' }}" dataBase="{{ current_period|date:'Y-m' }}" tipoRemessa="{{ submission_status }}">
    <contas>
        {% for account in accounts_with_movements %}
        <conta codigoConta="{{ account.operation_route.code }}" saldo="{{ account.balance.available|floatformat:2 }}"/>
        {% endfor %}
    </contas>
</documento>
```

### XML Template - Financial Report

```xml
<AnalyticalReport>
    <Organization>{{ midaz_onboarding.legal_name }} - Tax ID: {{ midaz_onboarding.legal_document }}</Organization>
    <GenerationDate>{% date_time "dd/MM/YYYY HH:mm" %}</GenerationDate>
    {%- with ledger = midaz_onboarding.ledger[0] %}
    <Ledger>{{ ledger.name }}</Ledger>

    {%- for account in midaz_onboarding.account %}
    <Account>
        <AccountID>{{ account.id }}</AccountID>
        <Alias>{{ account.alias }}</Alias>
        {%- with balance = filter(midaz_transaction.balance, "account_id", account.id)[0] %}
        <CurrentBalance>{{ balance.available }}</CurrentBalance>
        {%- endwith %}
        <Currency>{{ account.asset_code }}</Currency>
        <Operations>
        {%- for operation in midaz_transaction.operation %}
        {%- if operation.account_id == account.id %}
            {%- set original_amount = operation.amount %}
            {%- set discount_amount = original_amount * 0.03 %}
            {%- set final_amount = original_amount - discount_amount %}
        <Operation>
                <OperationID>{{ operation.id }}</OperationID>
                <Description>{{ operation.description }}</Description>
                <Type>{{ operation.type }}</Type>
                <OriginalAmount>{{ original_amount }}</OriginalAmount>
                <DiscountAmount>{{ discount_amount }}</DiscountAmount>
                <FinalAmount>{{ final_amount }}</FinalAmount>
                <Status>{{ operation.status }}</Status>
            </Operation>
        {%- endif %}
        {%- endfor %}
        </Operations>
        <AccountSummary>
            <TotalOperations>{% count_by midaz_transaction.operation if account_id == account.id %}</TotalOperations>
            <SumOfOperations>{% sum_by midaz_transaction.operation by "amount" if account_id == account.id %}</SumOfOperations>
            <AverageOfOperations>{% avg_by midaz_transaction.operation by "amount" if account_id == account.id %}</AverageOfOperations>
        </AccountSummary>
    </Account>
    {%- endfor %}
    {%- endwith %}
</AnalyticalReport>
```

Key XML Concepts:
- `{%-` removes extra whitespace in output
- `with` creates temporary variables for cleaner code
- `set` defines variables for calculations
- Array access: `ledger[0]` gets first item
- `filter()` function finds matching items

### HTML Template

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Financial Report</title>
  <style>
    table { width: 100%; border-collapse: collapse; }
    th, td { padding: 8px; text-align: left; border: 1px solid #ddd; }
  </style>
</head>
<body>
  <h1>Financial Report</h1>
  <p><strong>Date:</strong> {% date_time "dd/MM/YYYY HH:mm" %}</p>
  <p><strong>Organization:</strong> {{ midaz_onboarding.organization.0.legal_name }}</p>
  <p><strong>Ledger:</strong> {{ midaz_onboarding.ledger.0.name }}</p>

  {% for account in midaz_onboarding.account %}
    {% with balance = filter(midaz_transaction.balance, "account_id", account.id)[0] %}
      <h2>Account: {{ account.alias }}</h2>
      <p>ID: {{ account.id }} | Balance: {{ balance.available|floatformat:2 }}</p>

      <table>
        <thead>
          <tr>
            <th>Operation ID</th>
            <th>Type</th>
            <th>Amount</th>
            <th>Description</th>
          </tr>
        </thead>
        <tbody>
          {% for operation in midaz_transaction.operation %}
            {% if operation.account_id == account.id %}
              <tr>
                <td>{{ operation.id }}</td>
                <td>{{ operation.type }}</td>
                <td>{{ operation.amount|floatformat:2 }}</td>
                <td>{{ operation.description }}</td>
              </tr>
            {% endif %}
          {% endfor %}
        </tbody>
      </table>
    {% endwith %}
  {% endfor %}
</body>
</html>
```

### TXT Template - Fixed Width

```text
==========================================
          PROOF OF PAYMENT
==========================================
Date: {% date_time "dd/MM/YYYY HH:mm" %}
Ledger: {{ midaz_onboarding.ledger.0.name }}

{%- for transaction in midaz_transaction.transaction %}
------------------------------------------
Transaction ID: {{ transaction.id }}
Date: {{ transaction.created_at|date:"d/m/Y" }}
Amount: {{ transaction.amount|floatformat:2 }}
Status: {{ transaction.status }}

Source Accounts:
{%- for operation in filter(midaz_transaction.operation, "transaction_id", transaction.id) %}
{%- if operation.type == "DEBIT" %}
  Alias: {{ operation.account_alias|ljust:"20" }}
  Debit: {{ operation.amount|floatformat:2|rjust:"15" }}
{%- endif %}
{%- endfor %}

Target Accounts:
{%- for operation in filter(midaz_transaction.operation, "transaction_id", transaction.id) %}
{%- if operation.type == "CREDIT" %}
  Alias: {{ operation.account_alias|ljust:"20" }}
  Credit: {{ operation.amount|floatformat:2|rjust:"15" }}
{%- endif %}
{%- endfor %}
{%- endfor %}
==========================================
```

## Advanced Filtering in Requests

Report requests support complex filters:

```json
{
  "filters": {
    "midaz_onboarding": {
      "account": {
        "id": { "eq": ["123", "456"] },
        "created_at": { "between": ["2023-01-01", "2023-01-31"] },
        "status": { "in": ["active", "pending"] }
      }
    }
  }
}
```

Supported operators:
- `eq`: Equal to
- `gt`, `gte`: Greater than (or equal)
- `lt`, `lte`: Less than (or equal)
- `between`: Value within range
- `in`, `nin`: Value in/not in list

## Quick Reference

| Need | Filter/Pattern | Example |
|------|---------------|---------|
| Money (2 decimals) | `floatformat:2` | `{{ amount \| floatformat:2 }}` |
| Date YYYYMMDD | `date:"Ymd"` | `{{ date \| date:"Ymd" }}` |
| Date YYYY-MM | `date:"Y-m"` | `{{ date \| date:"Y-m" }}` |
| Date DD/MM/YYYY | `date:"d/m/Y"` | `{{ date \| date:"d/m/Y" }}` |
| CNPJ base (8 digits) | `slice:":8"` | `{{ cnpj \| slice:":8" }}` |
| Pad zeros left | `rjust:"10"` | `{{ code \| rjust:"10" }}` |
| Pad spaces right | `ljust:"20"` | `{{ name \| ljust:"20" }}` |
| Conditional value | `if-else` | `{% if x %}A{% else %}B{% endif %}` |
| Loop records | `for` | `{% for item in collection %}...{% endfor %}` |
| Remove whitespace | `{%-` prefix | `{%- for item in list %}` |
| Set variable | `set` | `{% set total = price * quantity %}` |
| Temporary variable | `with` | `{% with balance = account.balance %}` |

## Best Practices

### Template Rules
1. **NO business logic** - Only formatting and presentation
2. **Match structure exactly** - Follow regulatory specification precisely
3. **Use `.tpl` extension** - Required for Reporter to process
4. **Field naming** - Use **snake_case** consistently (e.g., `legal_document`, `opening_date`)

### Common Patterns
```django
# Validate required fields
{% if not organization.legal_document %}
  <!-- ERROR: Missing required CNPJ -->
{% endif %}

# Access first item in list
{{ collection.0.field }}

# Filter and get first match
{{ filter(collection, "field", value)[0].property }}

# Calculate inline
{% calc (value1 + value2) * 0.1 %}

# Format with multiple filters
{{ value|floatformat:2|rjust:"10" }}
```

### Testing Checklist

Before deployment:
1. [ ] File has `.tpl` extension
2. [ ] All mandatory fields present
3. [ ] Date formats match requirements
4. [ ] Numeric formats correct (decimals)
5. [ ] Structure matches specification
6. [ ] Conditionals handle all cases
7. [ ] No hardcoded values
8. [ ] Template under 100 lines
9. [ ] Loops use correct source
10. [ ] Special characters escaped

### Common Pitfalls
- Extra spaces in XML tags break parsing
- Missing XML encoding declaration
- Wrong decimal places for monetary values
- Incorrect date format for regulatory compliance
- Using camelCase instead of snake_case
- Forgetting array index `[0]` for single items
- Not using `{%-` to remove unwanted whitespace

## Data Contract Examples

### Expected structure from backend:
```python
{
    "organization": {
        "legal_document": "12345678000190",  # 14-digit CNPJ
        "legal_name": "Company Name"
    },
    "current_period": "2025-11-30",  # Date or ISO string
    "submission_status": "I",  # "I" or "S"
    "accounts_with_movements": [
        {
            "operation_route": {
                "code": "1.1.1.00.00"  # COSIF code
            },
            "balance": {
                "available": 1234567.89  # Decimal
            }
        }
    ]
}
```

## Template Output Control

Reporter renders templates **AS-IS**. The template structure **IS** the output structure. Keep templates minimal, clean, and focused only on data presentation.

---

*Technical reference extracted from Lerian Reporter documentation*