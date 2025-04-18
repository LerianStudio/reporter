{%- if not transaction_id -%}
{% set transaction_id = "019649e7-0166-7d94-8a4c-c9016d8b2a16" %}
{%- endif -%}
{%- for t in transaction.transaction -%}
{%- if transaction_id == "" or t.id == transaction_id -%}
##########################################
#        COMPROVANTE DE TRANSAÇÃO        #
##########################################

Transação ID: {{ t.id }}
Descrição: {{ t.description }}
Data de Criação: {{ t.created_at }}
Template: {{ t.template }}
Status: {{ t.status }}
Valor: {{ t.amount|floatformat:2 }}
Escala: {{ t.amount_scale }}
Moeda: {{ t.asset_code }}
Plano de Contas: {{ t.chart_of_accounts_group_name }}
{% endif %}
{% endfor %}

------------------------------------------
# Dados da Organização
------------------------------------------
{% for org in onboarding.organization %}
Nome Legal: {{ org.legal_name }}
Nome Fantasia: {{ org.doing_business_as }}
CNPJ: {{ org.legal_document }}
Endereço: {{ org.address }}
{% endfor %}

------------------------------------------
# Dados do Ledger
------------------------------------------
{% for l in onboarding.ledger %}
Ledger: {{ l.name }}
Status: {{ l.status }}
{% endfor %}

------------------------------------------
# Ativo
------------------------------------------
{% for a in onboarding.asset %}
Ativo: {{ a.name }}
Tipo: {{ a.type }}
Código: {{ a.code }}
{% endfor %}

------------------------------------------
# Contas Envolvidas na Operação
------------------------------------------
{% for operation in transaction.operation -%}
{% if operation.transaction_id == transaction_id %}
Operação ID: {{ operation.id }}
Descrição: {{ operation.description }}
Tipo: {{ operation.type }}
Conta: {{ operation.account_alias }}
Valor: {{ operation.amount|floatformat:2 }}
Saldo Disponível Após: {{ operation.available_balance_after|floatformat:2 }}
------------------------------------------
{% endif %}
{%- endfor %}

##########################################
#         FIM DO COMPROVANTE             #
##########################################