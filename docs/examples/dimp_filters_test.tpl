|0000|{{ midaz_onboarding.organization.0.legal_document|replace:".:"|replace:"/:"|replace:"-:" }}|{{ midaz_onboarding.organization.0.legal_document }}|
{% for acc in midaz_onboarding.account %}|1100|{% counter "1100" %}SP|{{ acc.id }}|{{ acc.alias|replace:"@:"|replace:"_:/" }}|
{% for tran in midaz_transaction.transaction %}{% if acc.organization_id == tran.organization_id %}|1101|{% counter "1101" %}{{ tran.id }}|
{% endif %}{% endfor %}{% endfor %}
|TOTAL_SP|{{ midaz_transaction.transaction|where:"status:APPROVED"|sum:"amount" }}|
|9900|1100|{% counter_show "1100" "1101" %}|
