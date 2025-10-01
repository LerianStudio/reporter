##########################################
#         COMPROVANTE DE PAGAMENTO       #
##########################################

Data da Geração: {% date_time "dd/MM/YYYY HH:mm" %}
Nome do Ledger: {{ midaz_onboarding.ledger.0.name }}

{%- for transaction in midaz_transaction.transaction %}
------------------------------------------
ID da Transação: {{ transaction.id }}
Data da Transação: {{ transaction.created_at }}
Status Transação: {{transaction.status}}
------------------------------------------

Contas de Origem:
{%- for operation in filter(midaz_transaction.operation, "transaction_id", transaction.id) %}
  {%- if operation.type == "DEBIT" %}
    - Alias: {{ operation.account_alias }}
  {%- endif %}
{%- endfor %}

Contas de Destino:
{%- for operation in filter(midaz_transaction.operation, "transaction_id", transaction.id) %}
  {%- if operation.type == "CREDIT" %}
    - Alias: {{ operation.account_alias }}
  {%- endif %}
{%- endfor %}

{%- endfor %}
------------------------------------------
Documento gerado automaticamente.