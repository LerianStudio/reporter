# Manual de Orientação do Leiaute da DIMP - Versão 10
## Declaração de Informações de Meios de Pagamentos

---

## 📋 APRESENTAÇÃO

Este manual orienta a geração do arquivo digital referente à **Declaração de Informações de Meios de Pagamentos – DIMP** prestadas por instituições financeiras e de pagamento, integrantes ou não do Sistema de Pagamentos Brasileiro – SPB.

### Escopo
- Transações mercantis com cartões de débito, crédito, de loja (private label) e demais instrumentos de pagamento eletrônico
- Informações prestadas por intermediadores (agenciadores, plataformas de delivery, marketplaces e similares)
- Transações comerciais ou de prestação de serviços intermediadas
- Base legal: **Convênio ICMS 134/16**

---

## 1. INFORMAÇÕES GERAIS

### Características do Sistema
- **Periodicidade**: Mensal
- **Formato**: Arquivo único para cada Unidade da Federação
- **Transmissão**: Via TED-TEF (validação, assinatura digital e transmissão)
- **Padronização**: Registros de transações padronizados

### Finalidades do Arquivo

| Código | Descrição |
|--------|-----------|
| 1 | **Normal**: Transações de pagamento e transferências regulares |
| 2 | **Substituto**: Retificação integral de informações do período |
| 3 | **Específico**: Cumprimento de notificação para CPF/CNPJ específico |
| 4 | **Arquivo Zerado**: Sem movimentação no período |
| 5 | **Encerramento**: Encerramento definitivo de atividades |
| 6 | **Autorização**: Para instituição parceira |

---

## 2. ESTRUTURA DO ARQUIVO

### Blocos Estruturantes

| Bloco | Descrição | Conteúdo |
|-------|-----------|----------|
| **0** | Abertura e Identificação | Dados das instituições e referências |
| **1** | Operações | Transações de pagamento e demais operações |
| **9** | Controle | Encerramento e totalização do arquivo |

### Hierarquia de Registros

```text
Nível 0    Nível 1    Nível 2    Nível 3    Nível 4    Nível 5
┌─────┐    ┌─────┐    ┌─────┐    ┌─────┐    ┌─────┐    ┌─────┐
│0000 │────│0001 │────│0002 │
│     │    │0990 │    │0005 │────│0006 │
│9999 │    │1001 │    │0100 │────│0105 │
          │1990 │    │0200 │────│0201 │
          │9001 │    │0300 │
          │9990 │    │0600 │
                     │0700 │
                     │1100 │────│1110 │────│1115 │────│1120 │
                     │1200 │────│1220 │
                     │1500 │
                     │1600 │
                     │9900 │
```

---

## 3. ESPECIFICAÇÕES TÉCNICAS

### Formato do Arquivo
- **Codificação**: ASCII - ISO 8859-1 (Latin-1)
- **Delimitador**: Pipe (|) - caractere 124 ASCII
- **Final de linha**: CR + LF (caracteres 13 e 10)
- **Tamanho máximo campos alfanuméricos**: 255 caracteres

### Tipos de Dados

| Formato | Sigla | Exemplo | Descrição |
|---------|-------|---------|-----------|
| Alfanumérico | X | Meio de Pagamento S/A | Texto livre |
| Numérico | N | 12 | Sem decimais |
| Numérico com decimal | N(2) | 12,25 | Com 2 casas decimais |
| Data | N | 20180122 | Formato aaaammdd |
| Hora | N | 081027 | Formato hhmmss |

---

## 4. REGISTROS PRINCIPAIS

### BLOCO 0 - Abertura e Identificação

#### Registro 0000 - Abertura do Arquivo
**Obrigatório** - Primeiro registro do arquivo

| Campo | Conteúdo | Tipo | Tamanho | Obr |
|-------|----------|------|---------|-----|
| REG | "0000" | N | 4 | O |
| COD_VER | Código da versão (10) | X | 2 | O |
| COD_FIN | Código da finalidade (1-6) | N | 1 | O |
| UF_FISCO | UF destinatária | X | 2 | O |
| CNPJ | CNPJ do declarante | X | 14 | O |
| NOME | Razão Social | X | - | O |
| DT_INI | Data inicial | N | 8 | O |
| DT_FIN | Data final | N | 8 | O |
| TP_AMB | Ambiente (1-prod, 2-homolog) | N | 1 | O |
| MES_VAL | Mês de validação (AAAAMM) | N | 6 | O |

#### Registro 0100 - Cadastro do Cliente
Identifica clientes/beneficiários de pagamento

| Campo | Conteúdo | Tipo | Tamanho | Obr |
|-------|----------|------|---------|-----|
| REG | "0100" | N | 4 | O |
| COD_CLIENTE | Código interno do cliente | X | - | O |
| CNPJ/CPF | Documento do beneficiário | X/N | 14/11 | OC |
| N_FANT | Nome fantasia | X | - | O |
| END | Endereço completo | X | - | O |
| CEP | CEP | N | 8 | O |
| COD_MUN | Código município IBGE | N | 7 | O |
| UF | Sigla UF | X | 2 | O |
| DT_CREDEN | Data credenciamento | N | 8 | O |

#### Registro 0200 - Meio de Captura
Identifica os meios/modos de captura

| Campo | Conteúdo | Tipo | Tamanho | Obr |
|-------|----------|------|---------|-----|
| REG | "0200" | N | 4 | O |
| COD_MCAPT | Código do meio de captura | X | - | O |
| NUM_LOG | Número lógico do terminal | X | - | O |
| TIPO_TECN | Tipo tecnologia (1-9) | N | 1 | O |
| TERM_PROP | Terminal próprio (0) ou terceiro (1) | N | 1 | O |
| MARCA | Marca/instituição | X | - | OC |
| SMARTPOS | É SmartPOS? (0-não, 1-sim) | N | 1 | OC |

##### Tipos de Tecnologia
- 1 - TEF-POS Integrados
- 2 - Mobile
- 3 - POS
- 4 - E-commerce
- 6 - URA/MOTO
- 7 - Pagamento em dinheiro/outra estrutura
- 8 - Conta titularidade individual
- 9 - Conta conjunta

---

### BLOCO 1 - Operações de Pagamento

#### Registro 1100 - Resumo Mensal das Operações
Totalização por cliente/período

| Campo | Conteúdo | Tipo | Tamanho | Dec | Obr |
|-------|----------|------|---------|-----|-----|
| REG | "1100" | N | 4 | - | O |
| COD_IP_PAR | Código instituição parceira | X | - | - | OC |
| COD_CLIENTE | Código do cliente | X | - | - | O |
| IND_COMEX | Transação internacional (0/1) | N | 1 | - | O |
| IND_EXTEMP | Extemporânea (0/1) | N | 1 | - | O |
| DT_INI | Data inicial | N | 8 | - | O |
| DT_FIN | Data final | N | 8 | - | O |
| VALOR | Valor total | N | 18 | 2 | O |
| QTD | Quantidade operações | N | 10 | - | O |

#### Registro 1115 - Operações por Comprovante
Detalhamento das transações

| Campo | Conteúdo | Tipo | Tamanho | Dec | Obr |
|-------|----------|------|---------|-----|-----|
| REG | "1115" | N | 4 | - | O |
| NSU | Número sequencial único | X | - | - | F |
| COD_AUT | Código autorização | X | - | - | O |
| ID_TRANSAC | ID da transação | X | - | - | O |
| IND_SPLIT | Operação splitada (0/1) | N | 1 | - | O |
| BANDEIRA | Código bandeira cartão | N | 2 | - | OC |
| HORA | Hora da transação | N | 6 | - | O |
| VALOR | Valor da operação | N | 18 | 2 | O |
| NAT_OPER | Natureza da operação (1-11) | N | 2 | - | O |

##### Naturezas de Operação
1. Cartão de Crédito
2. Cartão de Débito
3. Boleto transações próprias
4. Transferência de Recursos
5. Pagamento dinheiro/outra estrutura
6. PIX
7. Voucher/cartão pré-pago
8. Saque/troco estabelecimento
11. Recepção pagamento boletos terceiros

#### Registro 1120 - Intermediador de Serviços
Exclusivo para marketplaces/intermediadores

| Campo | Conteúdo | Tipo | Tamanho | Obr |
|-------|----------|------|---------|-----|
| REG | "1120" | N | 4 | O |
| UF_DEST | UF destino transação | X | 2 | O |
| CHAVE_NF | Chave NFe/NFCe | N | 44 | F |
| CNPJ/CPF_DEST | Documento destinatário | X/N | 14/11 | OC |
| UF_ORIGEM | UF origem mercadoria | X | 2 | OC |
| CNPJ/CPF_ORIGEM | Documento origem | X/N | 14/11 | OC |
| ID_PEDIDO | ID do pedido | X | - | O |
| ID_SELLER | ID vendedor na plataforma | X | - | O |

---

## 5. LIMITES E RESTRIÇÕES

### Limites Mínimos para Pessoa Física
- **Valor mensal**: R$ 3.375,00
- **Quantidade**: 30 transações
- **Regra**: Ambos os critérios devem ser atendidos

### Transações Excluídas do Envio

#### Transferências e Movimentações
- Mesma titularidade ou grupo empresarial (mesmo CNPJ raiz)
- Benefícios previdenciários e pensões (PF)
- Natureza salarial (PF)
- Benefícios sociais de órgãos públicos

#### Operações Financeiras
- Empréstimos e financiamentos
- Liquidação via SLC/CIP
- Operações de câmbio registradas no BACEN
- Resgate/pagamento de aplicações financeiras

#### Contas Específicas
- Instituições financeiras, seguradoras, previdência
- Condomínios
- Conta salário
- Cashbacks e programas de incentivo

#### Outros
- Crédito/restituição de IR
- Indenizações judiciais/seguros
- Doações e patrocínios (leis de incentivo)
- Créditos de estornos
- Depósitos em cheque ou dinheiro

---

## 6. VALIDAÇÕES E REGRAS

### Regras de Envio
- Um arquivo por UF por mês
- Transações da UF do beneficiário
- Exceção: finalidade 3 (intimação) ou registro 1120 com UF diferente

### Arquivos de Retificação
- Substituem integralmente o período
- Não aceitos registros extemporâneos
- Mantidas apenas transações extemporâneas do arquivo original

### Registros de Cancelamento
- Apenas em arquivos com finalidade "Normal" (1)
- Registros 1200 (detalhado) ou 1600 (consolidado)
- Deve informar data original e data do cancelamento

---

## 7. CONTROLES DO ARQUIVO

### Bloco 9 - Totalização

#### Registro 9900 - Totais por Tipo
Lista todos os tipos de registro com suas quantidades

#### Registro 9999 - Encerramento
Último registro com total de linhas do arquivo

### Validações Automáticas
- Contagem de linhas por bloco
- Somatórios de valores e quantidades
- Verificação de campos obrigatórios
- Validação de CNPJs/CPFs
- Conferência de dígitos verificadores

---

## 📅 CRONOGRAMA DE IMPLEMENTAÇÃO

### Versão 10
- **Instituída por**: Ato COTEPE
- **Obrigatoriedade**: 01/06/2025
- **Principais mudanças**:
  - Eliminação das finalidades 7 e 8
  - Novos campos para PIX
  - Ajustes em limites e validações

---

## 📞 SUPORTE E INFORMAÇÕES

Para dúvidas sobre o preenchimento e envio da DIMP, consulte:
- Manual completo disponível no site do CONFAZ
- Sistema TED-TEF para validação e transmissão
- Suporte técnico das Secretarias de Fazenda estaduais

---

*Este documento é um resumo do Manual de Orientação do Leiaute da DIMP v10.*
*Para informações completas, consulte o manual oficial.*