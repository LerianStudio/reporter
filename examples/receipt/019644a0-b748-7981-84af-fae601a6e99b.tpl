{%- for t in midaz_transaction.transaction -%}
##########################################
#        COMPROVANTE DE TRANSAÇÃO        #
##########################################

Transação ID: {{ t.id }}
Descrição: {{ t.description }}
Data de Criação: {{ t.created_at }}
Template: {{ t.template }}
Status: {{ t.status }}
Valor: {{ t.amount|scale:2 }}
Escala: {{ t.amount_scale }}
Moeda: {{ t.asset_code }}
Plano de Contas: {{ t.chart_of_accounts_group_name }}
{% endfor %}

------------------------------------------
# Dados da Organização
------------------------------------------
{% for org in midaz_onboarding.organization %}
Nome Legal: {{ org.legal_name }}
Nome Fantasia: {{ org.doing_business_as }}
CNPJ: {{ org.legal_document }}
Endereço: {{ org.address.line1 }}, {{ org.address.city }} - {{ org.address.state }}
{% endfor %}

------------------------------------------
# Dados do Ledger
------------------------------------------
{% for l in midaz_onboarding.ledger %}
Ledger: {{ l.name }}
Status: {{ l.status }}
{% endfor %}

------------------------------------------
# Ativo
------------------------------------------
{% for a in midaz_onboarding.asset %}
Ativo: {{ a.name }}
Tipo: {{ a.type }}
Código: {{ a.code }}
{% endfor %}

------------------------------------------
# Contas Envolvidas na Operação
------------------------------------------
{% for operation in midaz_transaction.operation -%}
Operação ID: {{ operation.id }}
Descrição: {{ operation.description }}
Tipo: {{ operation.type }}
Conta: {{ operation.account_alias }}
Valor: {{ operation.amount|scale:2 }}
Saldo Disponível Após: {{ operation.available_balance_after|scale:2 }}
------------------------------------------
{% endfor %}

##########################################
#         FIM DO COMPROVANTE             #
##########################################
