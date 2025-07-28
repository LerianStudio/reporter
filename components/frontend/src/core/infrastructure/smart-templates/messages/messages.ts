export const smartTemplatesApiMessages = {
  // Template validation and processing errors
  TPL_0001: {
    id: 'error.smartTemplates.missingRequiredFields',
    defaultMessage:
      'One or more required fields are missing. Please ensure all required fields like `description`, `template`, and `outputFormat` are included.'
  },
  TPL_0002: {
    id: 'error.smartTemplates.invalidFileFormat',
    defaultMessage:
      'The uploaded file must be a `.tpl` file. Other formats are not supported.'
  },
  TPL_0003: {
    id: 'error.smartTemplates.invalidOutputFormat',
    defaultMessage:
      'The outputFormat field must be one of: `html`, `csv`, or `xml`.'
  },
  TPL_0004: {
    id: 'error.smartTemplates.invalidHeader',
    defaultMessage:
      'One or more header values are missing or incorrectly formatted. Please verify all required headers.'
  },
  TPL_0005: {
    id: 'error.smartTemplates.invalidFileUploaded',
    defaultMessage:
      'The file you submitted is invalid. Please check the uploaded file for errors.'
  },
  TPL_0006: {
    id: 'error.smartTemplates.errorFileEmpty',
    defaultMessage:
      'The file you submitted is empty. Please check the uploaded file'
  },
  TPL_0007: {
    id: 'error.smartTemplates.errorFileContentInvalid',
    defaultMessage:
      'The file content is invalid because it is not in the expected format. Please check the uploaded file.'
  },
  TPL_0008: {
    id: 'error.smartTemplates.invalidMapFields',
    defaultMessage:
      'The field on template file is invalid. Please check the field mappings in your template.'
  },
  TPL_0009: {
    id: 'error.smartTemplates.invalidPathParameter',
    defaultMessage:
      'Path parameters are in an incorrect format. Please check the parameters and ensure they meet the required format before trying again.'
  },
  TPL_0010: {
    id: 'error.smartTemplates.updateOutputFormatWithoutTemplateFile',
    defaultMessage:
      'Can not update output format without passing template file. Please check information passed and try again.'
  },
  TPL_0011: {
    id: 'error.smartTemplates.entityNotFound',
    defaultMessage:
      'No entity was found for the given ID. Please make sure to use the correct ID for the entity you are trying to manage.'
  },
  TPL_0012: {
    id: 'error.smartTemplates.invalidTemplateID',
    defaultMessage:
      'The specified templateID is not a valid UUID. Please check the value passed.'
  },

  TPL_0014: {
    id: 'error.smartTemplates.missingRequiredFieldsInSchema',
    defaultMessage:
      'The fields mapped on template file are missing on tables schema. Please check the fields passed.'
  },
  TPL_0015: {
    id: 'error.smartTemplates.unexpectedFieldsInRequest',
    defaultMessage:
      'The request body contains more fields than expected. Please send only the allowed fields as per the documentation. The unexpected fields are listed in the fields object.'
  },
  TPL_0016: {
    id: 'error.smartTemplates.missingFieldsInRequest',
    defaultMessage:
      'Your request is missing one or more required fields. Please refer to the documentation to ensure all necessary fields are included in your request.'
  },
  TPL_0017: {
    id: 'error.smartTemplates.badRequest',
    defaultMessage:
      'The server could not understand the request due to malformed syntax. Please check the listed fields and try again.'
  },
  TPL_0018: {
    id: 'error.smartTemplates.internalServerError',
    defaultMessage:
      'The server encountered an unexpected error. Please try again later or contact support.'
  },
  TPL_0019: {
    id: 'error.smartTemplates.invalidQueryParameter',
    defaultMessage:
      'One or more query parameters are in an incorrect format. Please check the parameters and ensure they meet the required format before trying again.'
  },
  TPL_0020: {
    id: 'error.smartTemplates.invalidDateFormat',
    defaultMessage:
      "The 'initialDate', 'finalDate', or both are in the incorrect format. Please use the 'yyyy-mm-dd' format and try again."
  },
  TPL_0021: {
    id: 'error.smartTemplates.invalidFinalDate',
    defaultMessage:
      "The 'finalDate' cannot be earlier than the 'initialDate'. Please verify the dates and try again."
  },
  TPL_0022: {
    id: 'error.smartTemplates.dateRangeExceedsLimit',
    defaultMessage:
      "The range between 'initialDate' and 'finalDate' exceeds the permitted limit. Please adjust the dates and try again"
  },
  TPL_0023: {
    id: 'error.smartTemplates.invalidDateRange',
    defaultMessage:
      "Both 'initialDate' and 'finalDate' fields are required and must be in the 'yyyy-mm-dd' format. Please provide valid dates and try again"
  },
  TPL_0024: {
    id: 'error.smartTemplates.paginationLimitExceeded',
    defaultMessage:
      'The pagination limit exceeds the maximum allowed items per page. Please verify the limit and try again.'
  },
  TPL_0025: {
    id: 'error.smartTemplates.invalidSortOrder',
    defaultMessage:
      "The 'sort_order' field must be 'asc' or 'desc'. Please provide a valid sort order and try again."
  },
  TPL_0026: {
    id: 'error.smartTemplates.metadataKeyLengthExceeded',
    defaultMessage:
      'The metadata key exceeds the maximum allowed length of characters. Please use a shorter value.'
  },
  TPL_0026_REPORT: {
    id: 'error.smartTemplates.reportNotFound',
    defaultMessage: 'No report was found for the provided ID.'
  },
  TPL_0027: {
    id: 'error.smartTemplates.metadataValueLengthExceeded',
    defaultMessage:
      'The metadata value exceeds the maximum allowed length of characters. Please use a shorter value.'
  },
  TPL_0028: {
    id: 'error.smartTemplates.invalidMetadataNesting',
    defaultMessage:
      'The metadata object cannot contain nested values. Please ensure that the values are not nested and try again.'
  },
  TPL_0029: {
    id: 'error.smartTemplates.reportStatusNotFinished',
    defaultMessage:
      'The Report is not ready to download. Report is processing yet.'
  },
  TPL_0030: {
    id: 'error.smartTemplates.missingSchemaTable',
    defaultMessage:
      'There is a schema table missing. Please check your template file passed.'
  }
}
