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