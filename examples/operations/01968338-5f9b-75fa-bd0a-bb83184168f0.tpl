{%- with transaction = midaz_transaction.transaction[0] -%}
{%- with org = midaz_onboarding.organization[0] -%}
{{ "OPEINTRA"|ljust:8 }} 29042025 {{ org.legal_document|slice_str:"0:8" }}{{ transaction.operation|stringformat:"%08d"|length }}
{%- endwith %}
{%- for operation in midaz_transaction.operation %}
{{ operation.status }}{{ operation.asset_code|ljust:12 }}{{ operation.amount|scale:operation.amount_scale|stringformat:"%015s" }}
{%- endfor %}
{%- endwith %}