{% for alias in plugin_crm.aliases %}
Document: {{ alias.document }}
Account ID: {{ alias.account_id }}
Account: {{ alias.banking_details.account }}
Branch: {{ alias.banking_details.branch }}
type: {{ alias.banking_details.type }}
Opening Date: {{ alias.banking_details.opening_date }}
IBAN: {{ alias.banking_details.iban }}
Country Code: {{ alias.banking_details.country_code }}
Bank ID: {{ alias.banking_details.bank_id }}
Type: {{ alias.type }}
---------------------------------------------------
{% endfor %}