<?xml version="1.0" encoding="UTF-8"?>
<CCSDOC xmlns="http://www.bcb.gov.br/ccs/ACCS001.xsd">
    <BCARQ>
    <IdentdEmissor>12345678</IdentdEmissor>
    <IdentdDestinatario>00000000</IdentdDestinatario>
    <NomArq>ACCS001</NomArq>
    <NumRemessaArq>12233444</NumRemessaArq>
</BCARQ>
<SISARQ>
    <CCSArqAtlzDiaria>
        <Repet_ACCS001_Pessoa>
{%- for alias in plugin_crm.aliases -%}
{%- for holder in plugin_crm.holders -%}
{%- if holder.document == alias.document %}
            <Grupo_ACCS001_Pessoa>
                <TpOpCCS>I</TpOpCCS>
                <QualifdrOpCCS>N</QualifdrOpCCS>
                <TpPessoa>{%- if holder.type == "NATURAL_PERSON" -%}F{%- else -%}J{%- endif -%}</TpPessoa>
                <CNPJ_CPFPessoa>{{holder.document}}</CNPJ_CPFPessoa>
                <DtIni>{{alias.banking_details.opening_date}}</DtIni>
                <DtFim></DtFim>
            </Grupo_ACCS001_Pessoa>
{%- endif -%}
{%- endfor %}
{%- endfor %}
        </Repet_ACCS001_Pessoa>
        <QtdOpCCS>{% count_by plugin_crm.aliases %}</QtdOpCCS>
        <DtMovto>{% date_time "YYYY-MM-DD" %}</DtMovto>
    </CCSArqAtlzDiaria>
</SISARQ>
</CCSDOC>
