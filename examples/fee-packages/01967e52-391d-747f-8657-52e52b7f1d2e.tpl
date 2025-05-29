{% for package in plugin_fees.package %}
Description: {{ package.description }}
Chart Of Accounts: {{ package.chart_of_account }}
Ledger: {{ package.ledger_id }}
---------------------------------------------------
{% endfor %}