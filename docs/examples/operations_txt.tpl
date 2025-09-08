{%- with transaction = midaz_transaction.transaction[0] -%}
{%- with org = midaz_onboarding.organization[0] -%}
{{ "OPEINTRA"|ljust:8 }} 29042025 {{ org.legal_document|slice_str:"0:8" }}
{%- endwith %}
{%- for operation in midaz_transaction.operation %}
{{ operation.status }}{{ operation.asset_code|ljust:12 }}{{ operation.amount|stringformat:"%015s" }}
{%- endfor %}
{%- endwith %}