import get from 'lodash/get'
import { IntlShape } from 'react-intl'
import {
  defaultErrorMap,
  ErrorMapCtx,
  ZodIssueCode,
  ZodIssueOptionalMessage,
  ZodParsedType
} from 'zod'
import messages from './messages'
import dayjs from 'dayjs'

/**
 * TODO: Proper implement this
 * @param intl
 * @returns
 */
export default function createZodMap(intl: IntlShape) {
  return (issue: ZodIssueOptionalMessage, ctx: ErrorMapCtx) => {
    let message: string = defaultErrorMap(issue, ctx).message

    switch (issue.code) {
      case ZodIssueCode.invalid_type:
        if (issue.received === ZodParsedType.undefined) {
          message = intl.formatMessage(messages.invalid_type_received_undefined)
        }
        break
      case ZodIssueCode.too_small:
        if (issue.message && issue.message in messages) {
          message = intl.formatMessage(
            messages[issue.message as keyof typeof messages]
          )
        } else {
          const minimum =
            issue.type === 'date'
              ? dayjs.unix(issue.minimum as number).format('LL')
              : issue.minimum
          const precisionMinimum = issue.exact
            ? 'exact'
            : issue.inclusive
              ? 'inclusive'
              : 'not_inclusive'
          const keyMinimum = `too_small_${issue.type}_${precisionMinimum}`

          if (!(keyMinimum in messages)) {
            throw new Error(
              `Zod Intl: ${keyMinimum} id is not defined in messages`
            )
          }

          // TODO: review this error
          // @ts-ignore
          message = intl.formatMessage(messages[keyMinimum], { minimum })
        }
        break
      case ZodIssueCode.invalid_enum_value:
        if (issue.message && issue.message in messages) {
          message = intl.formatMessage(
            messages[issue.message as keyof typeof messages]
          )
          break
        }
        message = intl.formatMessage(messages.invalid_type_received_undefined)
        break
      case ZodIssueCode.too_big:
        const maximum =
          issue.type === 'date'
            ? dayjs.unix(issue.maximum as number).format('LL')
            : issue.maximum
        const precisionMaximum = issue.exact
          ? 'exact'
          : issue.inclusive
            ? 'inclusive'
            : 'not_inclusive'
        const keyMaximum = `too_big_${issue.type}_${precisionMaximum}`

        if (!(keyMaximum in messages)) {
          throw new Error(
            `Zod Intl: ${keyMaximum} id is not defined in messages`
          )
        }

        // @ts-ignore - formatMessage can return ReactNode[] with values
        message = intl.formatMessage(
          messages[keyMaximum as keyof typeof messages],
          { maximum }
        )
        break
      case ZodIssueCode.custom:
        if (issue?.params?.id) {
          if (!(issue.params.id in messages)) {
            throw new Error(
              `Zod Intl: ${issue.params.id} id is not defined in messages`
            )
          }
          message = intl.formatMessage(
            get(messages, issue.params.id),
            issue.params
          )
        }
        break
    }
    return { message }
  }
}
