import { defineMessages } from '@/lib/intl'

export const reporterApiMessages = defineMessages({
  // Template validation and processing errors
  'TPL-0001': {
    id: 'error.reporter.missingRequiredFields',
    defaultMessage:
      'One or more required fields are missing. Please ensure all required fields like `description`, `template`, and `outputFormat` are included.'
  },
  'TPL-0002': {
    id: 'error.reporter.invalidFileFormat',
    defaultMessage:
      'The uploaded file must be a `.tpl` file. Other formats are not supported.'
  },
  'TPL-0003': {
    id: 'error.reporter.invalidOutputFormat',
    defaultMessage:
      'The outputFormat field must be one of: `html`, `csv`, or `xml`.'
  },
  'TPL-0004': {
    id: 'error.reporter.invalidHeader',
    defaultMessage:
      'One or more header values are missing or incorrectly formatted. Please verify all required headers.'
  },
  'TPL-0005': {
    id: 'error.reporter.invalidFileUploaded',
    defaultMessage:
      'The file you submitted is invalid. Please check the uploaded file for errors.'
  },
  'TPL-0006': {
    id: 'error.reporter.errorFileEmpty',
    defaultMessage:
      'The file you submitted is empty. Please check the uploaded file'
  },
  'TPL-0007': {
    id: 'error.reporter.errorFileContentInvalid',
    defaultMessage:
      'The file content is invalid because it is not in the expected format. Please check the uploaded file.'
  },
  'TPL-0008': {
    id: 'error.reporter.invalidMapFields',
    defaultMessage:
      'The field on template file is invalid. Please check the field mappings in your template.'
  },
  'TPL-0009': {
    id: 'error.reporter.invalidPathParameter',
    defaultMessage:
      'Path parameters are in an incorrect format. Please check the parameters and ensure they meet the required format before trying again.'
  },
  'TPL-0010': {
    id: 'error.reporter.updateOutputFormatWithoutTemplateFile',
    defaultMessage:
      'Can not update output format without passing template file. Please check information passed and try again.'
  },
  'TPL-0011': {
    id: 'error.reporter.entityNotFound',
    defaultMessage:
      'No entity was found for the given ID. Please make sure to use the correct ID for the entity you are trying to manage.'
  },
  'TPL-0012': {
    id: 'error.reporter.invalidTemplateID',
    defaultMessage:
      'The specified templateID is not a valid UUID. Please check the value passed.'
  },

  'TPL-0014': {
    id: 'error.reporter.missingRequiredFieldsInSchema',
    defaultMessage:
      'The fields mapped on template file are missing on tables schema. Please check the fields passed.'
  },
  'TPL-0015': {
    id: 'error.reporter.unexpectedFieldsInRequest',
    defaultMessage:
      'The request body contains more fields than expected. Please send only the allowed fields as per the documentation. The unexpected fields are listed in the fields object.'
  },
  'TPL-0016': {
    id: 'error.reporter.missingFieldsInRequest',
    defaultMessage:
      'Your request is missing one or more required fields. Please refer to the documentation to ensure all necessary fields are included in your request.'
  },
  'TPL-0017': {
    id: 'error.reporter.badRequest',
    defaultMessage:
      'The server could not understand the request due to malformed syntax. Please check the listed fields and try again.'
  },
  'TPL-0018': {
    id: 'error.reporter.internalServerError',
    defaultMessage:
      'The server encountered an unexpected error. Please try again later or contact support.'
  },
  'TPL-0019': {
    id: 'error.reporter.invalidQueryParameter',
    defaultMessage:
      'One or more query parameters are in an incorrect format. Please check the parameters and ensure they meet the required format before trying again.'
  },
  'TPL-0020': {
    id: 'error.reporter.invalidDateFormat',
    defaultMessage:
      "The 'initialDate', 'finalDate', or both are in the incorrect format. Please use the 'yyyy-mm-dd' format and try again."
  },
  'TPL-0021': {
    id: 'error.reporter.invalidFinalDate',
    defaultMessage:
      "The 'finalDate' cannot be earlier than the 'initialDate'. Please verify the dates and try again."
  },
  'TPL-0022': {
    id: 'error.reporter.dateRangeExceedsLimit',
    defaultMessage:
      "The range between 'initialDate' and 'finalDate' exceeds the permitted limit. Please adjust the dates and try again"
  },
  'TPL-0023': {
    id: 'error.reporter.invalidDateRange',
    defaultMessage:
      "Both 'initialDate' and 'finalDate' fields are required and must be in the 'yyyy-mm-dd' format. Please provide valid dates and try again"
  },
  'TPL-0024': {
    id: 'error.reporter.paginationLimitExceeded',
    defaultMessage:
      'The pagination limit exceeds the maximum allowed items per page. Please verify the limit and try again.'
  },
  'TPL-0025': {
    id: 'error.reporter.invalidSortOrder',
    defaultMessage:
      "The 'sort_order' field must be 'asc' or 'desc'. Please provide a valid sort order and try again."
  },
  'TPL-0026': {
    id: 'error.reporter.metadataKeyLengthExceeded',
    defaultMessage:
      'The metadata key exceeds the maximum allowed length of characters. Please use a shorter value.'
  },
  'TPL-0026-REPORT': {
    id: 'error.reporter.reportNotFound',
    defaultMessage: 'No report was found for the provided ID.'
  },
  'TPL-0027': {
    id: 'error.reporter.metadataValueLengthExceeded',
    defaultMessage:
      'The metadata value exceeds the maximum allowed length of characters. Please use a shorter value.'
  },
  'TPL-0028': {
    id: 'error.reporter.invalidMetadataNesting',
    defaultMessage:
      'The metadata object cannot contain nested values. Please ensure that the values are not nested and try again.'
  },
  'TPL-0029': {
    id: 'error.reporter.reportStatusNotFinished',
    defaultMessage:
      'The Report is not ready to download. Report is processing yet.'
  },
  'TPL-0030': {
    id: 'error.reporter.missingSchemaTable',
    defaultMessage:
      'There is a schema table missing. Please check your template file passed.'
  }
})
