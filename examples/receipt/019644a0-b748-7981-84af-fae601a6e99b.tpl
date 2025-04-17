##########################################
#        COMPROVANTE DE TRANSAÇÃO        #
##########################################

Transação ID: {{ transaction.transaction.id }}
Descrição: {{ transaction.transaction.description }}
Data de Criação: {{ transaction.transaction.created_at }}
Template: {{ transaction.transaction.template }}
Status: {{ transaction.transaction.status }}
Valor: {{ transaction.transaction.amount|floatformat:2 }}
Escala: {{ transaction.transaction.amount_scale }}
Moeda: {{ transaction.transaction.asset_code }}
Plano de Contas: {{ transaction.transaction.chart_of_accounts_group_name }}

------------------------------------------
# Dados da Organização
------------------------------------------
Nome Legal: {{ onboarding.organization.legal_name }}
Nome Fantasia: {{ onboarding.organization.doing_business_as }}
CNPJ: {{ onboarding.organization.legal_document }}
Endereço: {{ onboarding.organization.address }}

------------------------------------------
# Dados do Ledger
------------------------------------------
Ledger: {{ onboarding.ledger.name }}
Status: {{ onboarding.ledger.status }}

------------------------------------------
# Ativo
------------------------------------------
Ativo: {{ onboarding.asset.name }}
Tipo: {{ onboarding.asset.type }}
Código: {{ onboarding.asset.code }}

------------------------------------------
# Contas Envolvidas na Operação
------------------------------------------

{% for operation in transaction.operation %}
Operação ID: {{ operation.id }}
Descrição: {{ operation.description }}
Tipo: {{ operation.type }}
Conta: {{ operation.account_alias }}
Valor: {{ operation.amount|floatformat:2 }}
Saldo Disponível Após: {{ operation.available_balance_after|floatformat:2 }}
------------------------------------------
{% endfor %}

##########################################
#         FIM DO COMPROVANTE             #
##########################################
