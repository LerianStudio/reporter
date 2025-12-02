---
name: regulatory-templates-gate3
description: Gate 3 of regulatory templates - generates the complete .tpl template file with all validated mappings
---

# Regulatory Templates - Gate 3: Template File Generation

## Overview

**This sub-skill executes Gate 3 of the regulatory template workflow: generating the complete .tpl template file with all validated mappings and transformations from Gates 1-2.**

**Parent skill:** `regulatory-templates`

**Prerequisites:**
- Gate 1 PASSED (field mappings complete)
- Gate 2 PASSED (validations confirmed)
- Context object with Gates 1-2 results

**Output:** Generated .tpl template file ready for use

---

## Foundational Principle

**Template generation is the final quality gate before production deployment.**

Gate 3 transforms validated specifications into production artifacts:
- **Agent-based generation**: finops-automation applies validated mappings consistently - manual creation introduces human error
- **Two-file separation**: Clean .tpl (production code) + .tpl.docs (documentation) - inline comments bloat production artifacts
- **All mandatory fields**: 100% inclusion required - 95% = 5% of regulatory data missing in BACEN submission
- **Correct transformations**: Django filters applied per Gates 1-2 validation - errors here multiply in every submission
- **Valid syntax**: Template must execute without errors - syntax failures block Reporter deployment

**Skipping requirements in Gate 3 means:**
- Manual creation bypasses systematic validation (fatigue errors, missed transformations)
- Single-file output mixes production code with documentation (maintenance nightmare)
- Missing fields cause BACEN submission failures (compliance violations)
- Invalid syntax blocks deployment (emergency fixes under pressure)

**Gate 3 is not automation for convenience - it's the final verification layer.**

---

## When to Use

**Called by:** `regulatory-templates` skill after Gate 2 passes

**Purpose:** Create the final Django/Jinja2 template file with all field mappings, transformations, and validation logic

---

## Generation Requirements

Gate 3 produces the final production artifact. These requirements ensure quality and regulatory compliance.

### Why Agent-Based Generation Matters

| Consideration | Manual Approach | Agent Approach |
|---------------|-----------------|----------------|
| **Consistency** | Varies with fatigue/experience | Systematic validation |
| **Traceability** | Ad-hoc decisions | Gates 1-2 mappings applied |
| **Completeness** | Risk of missing fields | All fields included |
| **Maintainability** | Mixed code/docs | Clean separation |

### Mandatory Requirements

**1. Agent-Based Generation**
- Use `finops-automation` agent for template generation
- Agent applies Gates 1-2 validations consistently
- Exceptions require documented justification and approval

**2. Two-File Output**
- Generate `.tpl` (clean template code) + `.tpl.docs` (documentation)
- Production artifacts remain clean and maintainable

**3. Complete Field Coverage**
- All mandatory fields must be included (100% coverage)
- Missing fields risk regulatory compliance failures

**4. Validated Output**
- Use agent output as-is unless corrections are documented
- Manual edits should be reviewed to prevent drift from validated specs

### Common Pitfalls to Avoid

| Shortcut | Risk | Recommendation |
|----------|------|----------------|
| Manual creation "to save time" | Inconsistent validation | Use agent workflow |
| Single file with inline docs | Maintenance burden | Separate `.tpl` and `.docs` |
| Partial field coverage | Compliance failures | Ensure 100% mandatory fields |
| Optimizing agent output | Drift from validated spec | Document any changes |

### Quality Checkpoint

Before finalizing, verify:
- ✅ All mandatory fields from Gates 1-2 are present
- ✅ Template and documentation are separate files
- ✅ Agent-generated output is used (or changes documented)

---

## Gate 3 Process

### Agent Dispatch

**Use the Task tool to dispatch the finops-automation agent for template generation:**

1. **Invoke the Task tool with these parameters:**
   - `subagent_type`: "finops-automation"
   - `model`: "sonnet"
   - `description`: "Gate 3: Generate template file"
   - `prompt`: Use the prompt template below with accumulated context from Gates 1-2

2. **Prompt Template for Gate 3:**

```text
GATE 3: TEMPLATE FILE GENERATION

CONTEXT FROM GATES 1-2:
- Template: [insert context.template_name]
- Template Code: [insert context.template_code]
- Authority: [insert context.authority]
- Fields Mapped: [insert context.field_mappings.length]
- Validation Rules: [insert context.validation_rules.length]

FIELD MAPPINGS FROM GATE 1:
[For each field in context.field_mappings, list:]
Field [field_code]: [field_name]
- Source: [selected_mapping]
- Transformation: [transformation or 'none']
- Confidence: [confidence_score]%
- Required: [required]

VALIDATION RULES FROM GATE 2:
[For each rule in context.validation_rules, list:]
Rule [rule_id]: [description]
- Formula: [formula]

TASKS:
1. Generate CLEAN .tpl file with ONLY Django/Jinja2 template code
2. Include all field mappings with transformations
3. Apply correct template syntax for Reporter
4. Structure according to regulatory format requirements
5. Include conditional logic where needed
6. Use MINIMAL inline comments (max 1-2 lines critical notes only)

CRITICAL - NAMING CONVENTION:
⚠️ ALL field names are in SNAKE_CASE standard
- Gate 1 has already converted all fields to snake_case
- Examples:
  * Use {{ legal_document }} (converted from legalDocument)
  * Use {{ operation_route }} (already snake_case)
  * Use {{ opening_date }} (converted from openingDate)
  * Use {{ natural_person }} (converted from naturalPerson)
- ALL fields follow snake_case convention consistently
- No conversion needed - fields arrive already standardized

CRITICAL - DATA SOURCES:
⚠️ ALWAYS prefix fields with the correct data source!
Reference: .claude/docs/regulatory/DATA_SOURCES.md

Available Data Sources:
- midaz_onboarding: organization, account (cadastral data)
- midaz_transaction: operation_route, balance, operation (transactional data)

Field Path Format: {data_source}.{entity}.{index?}.{field}

Examples:
- CNPJ: {{ midaz_onboarding.organization.0.legal_document|slice:':8' }}
- COSIF: {{ midaz_transaction.operation_route.code }}
- Saldo: {{ midaz_transaction.balance.available }}
- Data: {% date_time "YYYY/MM" %}

WRONG: {{ organization.legal_document }}
CORRECT: {{ midaz_onboarding.organization.0.legal_document }}

TEMPLATE STRUCTURE:
- Use proper hierarchy per regulatory spec
- Include loops for repeating elements (accounts, transactions)
- Apply transformations using Django filters
- NO DOCUMENTATION BLOCKS - only functional template code

OUTPUT FILES (Generate TWO separate files):

FILE 1 - TEMPLATE (CLEAN):
- Filename: [template_code]_preview.tpl
- Content: ONLY the Django/Jinja2 template code
- NO extensive comments or documentation blocks
- Maximum 1-2 critical inline comments if absolutely necessary
- Ready for DIRECT use in Reporter without editing

FILE 2 - DOCUMENTATION:
- Filename: [template_code]_preview.tpl.docs
- Content: Full documentation including:
  * Field mappings table
  * Transformation rules
  * Validation checklist
  * Troubleshooting guide
  * Maintenance notes
  * All the helpful documentation

CRITICAL: The .tpl file must be production-ready and contain ONLY
the functional template code. All documentation goes in .docs file.

COMPLETION STATUS must be COMPLETE or INCOMPLETE.
```

3. **Complete Example - CADOC 4010:**

```javascript
// Task tool invocation with substituted values
Task({
  subagent_type: "finops-automation",
  model: "sonnet",
  description: "Gate 3: Generate CADOC 4010 template",
  prompt: `
GATE 3: TEMPLATE FILE GENERATION

CONTEXT FROM GATES 1-2:
- Template: CADOC 4010 - Informações de Cadastro
- Template Code: 4010
- Authority: BACEN
- Fields Mapped: 47
- Validation Rules: 12

FIELD MAPPINGS FROM GATE 1:
Field CNPJ_BASE: CNPJ da Instituição (8 dígitos)
- Source: midaz_onboarding.organization.0.legal_document
- Transformation: slice:':8'
- Confidence: 95%
- Required: true

Field DATA_BASE: Data Base do Documento
- Source: report_period.reference_date
- Transformation: date:'Y-m'
- Confidence: 98%
- Required: true

[... remaining 45 fields ...]

VALIDATION RULES FROM GATE 2:
Rule V001: CNPJ deve ter 8 dígitos
- Formula: len(CNPJ_BASE) == 8

[... remaining rules ...]

TASKS:
1. Generate .tpl file with all 47 field mappings
2. Generate .tpl.docs with field documentation
3. Apply all transformations from Gate 1
4. Include validation rules from Gate 2

OUTPUT FILES:
- cadoc4010_20251202_preview.tpl
- cadoc4010_20251202_preview.tpl.docs

COMPLETION STATUS must be COMPLETE or INCOMPLETE.
`
})
```

---

## Agent Execution

The agent `finops-automation` will handle all technical aspects:

- Analyze template requirements based on authority type
- Generate appropriate template structure (XML, JSON, etc.)
- Apply all necessary transformations using Django/Jinja2 filters
- Include conditional logic for business rules
- Ensure compliance with regulatory format specifications

---

## Expected Output

The agent will generate two files:

1. **Template File**: `{template_code}_{timestamp}_preview.tpl`
   - Contains the functional Django/Jinja2 template code
   - Ready for direct use in Reporter
   - Minimal inline comments only if necessary

2. **Documentation File**: `{template_code}_{timestamp}_preview.tpl.docs`
   - Contains full documentation
   - Field mapping details
   - Maintenance notes

---

## Red Flags - STOP Immediately

If you catch yourself thinking ANY of these, STOP and re-read the NO EXCEPTIONS section:

### Manual Shortcuts
- "Create .tpl manually, faster"
- "Edit agent output for optimization"
- "I can write cleaner code"
- "Agent is too verbose"

### File Structure Violations
- "One file easier to maintain"
- "Inline comments instead of .docs"
- "Merge documentation into .tpl"
- "Two files is over-engineering"

### Partial Completion
- "45/47 fields works for most cases"
- "Skip edge case fields"
- "Add missing fields later"
- "99% is good enough"

### Justification Language
- "Being pragmatic"
- "I'm too tired for agent wait"
- "Manual is faster"
- "Over-engineering"
- "Optimization is better"

### If You See These Red Flags

1. **Acknowledge rationalization** ("I'm trying to skip agent generation")
2. **Read NO EXCEPTIONS** (understand why agent is required)
3. **Read Rationalization Table** (see excuse refuted)
4. **Use agent completely** (no manual shortcuts)

**Template generation shortcuts waste all Gates 1-2 validation work.**

---

## Pass/Fail Criteria

### PASS Criteria
- ✅ Template file generated successfully
- ✅ All mandatory fields included
- ✅ Transformations correctly applied
- ✅ Django/Jinja2 syntax valid
- ✅ Output format matches specification
- ✅ File saved with correct extension

### FAIL Criteria
- ❌ Missing mandatory fields
- ❌ Invalid template syntax
- ❌ Transformation errors
- ❌ File generation failed

---

## State Tracking

### After PASS:

```yaml
SKILL: regulatory-templates-gate3
GATE: 3 - Template File Generation
STATUS: PASSED ✅
TEMPLATE: {context.template_selected}
FILE: {filename}
FIELDS: {fields_included}/{total_fields}
NEXT: Template ready for use
EVIDENCE: File generated successfully
BLOCKERS: None
```

### After FAIL:

```yaml
SKILL: regulatory-templates-gate3
GATE: 3 - Template File Generation
STATUS: FAILED ❌
TEMPLATE: {context.template_selected}
ERROR: {error_message}
NEXT: Fix generation issues
EVIDENCE: {specific_failure}
BLOCKERS: {blocker_description}
```

---

## Output to Parent Skill

Return to `regulatory-templates` main skill:

```javascript
{
  "gate3_passed": true/false,
  "template_file": {
    "filename": "cadoc4010_20251119_preview.tpl",
    "path": "/path/to/file",
    "size_bytes": 2048,
    "fields_included": 9
  },
  "ready_for_use": true/false,
  "next_action": "template_complete" | "fix_and_regenerate"
}
```

---

## Common Template Patterns

### Field Naming (snake_case)

All fields from Gate 1 are provided in snake_case format:

```django
{{ organization.legal_document }}    # ✅ Correct
{{ account.opening_date }}           # ✅ Correct
{{ organization.legalDocument }}     # ❌ Wrong (camelCase)
```

### Iterating Collections
```django
{% for item in collection %}
    {{ item.field }}
{% endfor %}
```

### Conditional Fields
```django
{% if condition %}
    <field>{{ value }}</field>
{% endif %}
```

### Nested Objects
```django
{{ parent.child.grandchild }}
```

### Filters Chain
```django
{{ value|slice:':8'|upper }}
```

---

## Remember

1. **Use exact field paths** from Gate 1 mappings (all fields are snake_case)
2. **Apply all transformations** validated in Gate 2
3. **Include comments** for maintainability
4. **Follow regulatory format** exactly
5. **Test syntax validity** before saving
6. **Document assumptions** made during generation