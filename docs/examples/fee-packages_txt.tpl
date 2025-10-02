{% for package in plugin_fees.package %}
Description: {{ package.description }}
Max Amount: {{ package.maximum_amount }}
Min Amount: {{ package.minimum_amount }}
Ledger: {{ package.ledger_id }}
---------------------------------------------------
{% endfor %}