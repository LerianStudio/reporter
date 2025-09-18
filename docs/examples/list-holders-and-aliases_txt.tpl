List of Aliases
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

List of Holders
{% for holder in plugin_crm.holders %}
Document: {{ holder.document }}
Type: {{ holder.type }}
Adress Primary Line 1: {{ holder.addresses.primary.line_1 }}
Adress Additional Line 1: {{ holder.addresses.additional_1.line_1 }}
Contact mobile Phone: {{ holder.contact.mobile_phone }}
Contact primary email: {{ holder.contact.primary_email }}
Trade Name: {{ holder.legal_person.trade_name }}
Representative Name: {{ holder.legal_person.representative.name }}
Natural Person Mother Name: {{ holder.natural_person.mother_name }}
Natural Person Father Name: {{ holder.natural_person.father_name }}
---------------------------------------------------
{% endfor %}