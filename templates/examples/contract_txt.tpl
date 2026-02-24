##########################################
#     CONTRATO DE ABERTURA DE CONTA      #
##########################################

CONTRATANTE:
Nome: {{ midaz_onboarding.organization.0.legal_name }}
CNPJ: {{ midaz_onboarding.organization.0.legal_document }}
Endereço: {{ midaz_onboarding.organization.0.address.line1 }}, {{ midaz_onboarding.organization.0.address.city }} - {{ midaz_onboarding.organization.0.address.state }}

DADOS DA CONTA:
Identificador da Conta: {{ midaz_onboarding.account.0.id }}
Alias da Conta: {{ midaz_onboarding.account.0.alias }}
Saldo Inicial: {{ midaz_transaction.balance.0.available }}
Moeda da Conta: {{ midaz_onboarding.account.0.asset_code }}

DADOS DO LEDGER ASSOCIADO:
Nome do Ledger: {{ midaz_onboarding.ledger.0.name }}

OBJETO DO CONTRATO:
O presente instrumento tem como objeto a abertura e manutenção da conta vinculada à organização supracitada, sob as condições estabelecidas neste contrato.

OBRIGAÇÕES:
- Manter saldo positivo para utilização de serviços.
- Observar as regras internas estabelecidas pela organização {{ midaz_onboarding.organization.0.legal_name }}.

CONDIÇÕES GERAIS:
- Este contrato é válido a partir da assinatura pelas partes.
- Quaisquer alterações no status do ledger poderão impactar os serviços relacionados à conta.

ASSINATURA:
________________________________________
Representante da CONTRATANTE
{{ midaz_onboarding.organization.0.legal_name }}