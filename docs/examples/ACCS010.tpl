<?xml version="1.0" encoding="UTF-8"?>
<CCSDOC xmlns="http://www.bcb.gov.br/ccs/ACCS010.xsd">
    <BCARQ>
        <IdentdEmissor>12345678</IdentdEmissor>
        <IdentdDestinatario>00000000</IdentdDestinatario>
        <NomArq>ACCS010</NomArq>
        <NumRemessaArq>{% date_time "yyyymmdd" %}0001</NumRemessaArq>
    </BCARQ>
    <SISARQ>
        <CCSArqTransRelac>
            <CNPJBaseNovRespons>12345679</CNPJBaseNovRespons>
            {%- for alias in plugin_crm.aliases %}
            {%- for holder in plugin_crm.holders %}
            {%- if alias.holder_id == holder._id and (holder.type == "legal_person" or holder.type == "natural_person") %}
            <Repet_ACCS010_Pessoa>
                <CNPJBasePart></CNPJBasePart>
                <TpPessoa>{% if holder.type == "natural_person" %}F{% else %}J{% endif %}</TpPessoa>
                <CNPJ_CPFPessoa>{{holder.document}}</CNPJ_CPFPessoa>
                <DtIni>{{alias.banking_details.opening_date}}</DtIni>
                <DtFim></DtFim>
            </Repet_ACCS010_Pessoa>
            {%- endif %}
            {%- endfor %}
            {%- endfor %}
            
            <QtdOpCCS>{% count_by plugin_crm.aliases %}</QtdOpCCS>
        </CCSArqTransRelac>
    </SISARQ>
</CCSDOC>
