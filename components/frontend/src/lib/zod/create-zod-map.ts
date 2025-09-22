import get from 'lodash/get'
import { IntlShape } from 'react-intl'
import { z } from 'zod'
import messages from './messages'
import dayjs from 'dayjs'

export default function createZodMap(intl: IntlShape) {
  return (issue: any, ctx?: { defaultError?: string }) => {
    let message: string = ctx?.defaultError || intl.formatMessage(messages.invalid_type)

    switch (issue.code) {
      case 'invalid_type':
        if ((issue as any).received === 'undefined') {
          message = intl.formatMessage(messages.invalid_type_received_undefined)
        } else {
          message = intl.formatMessage(messages.invalid_type)
        }
        break
      case 'too_small':
        const minimum =
          (issue as any).origin === 'date'
            ? dayjs.unix((issue as any).minimum as number).format('LL')
            : (issue as any).minimum
        const precisionMinimum = (issue as any).exact
          ? 'exact'
          : (issue as any).inclusive
            ? 'inclusive'
            : 'not_inclusive'
        const keyMinimum = `too_small_${(issue as any).origin}_${precisionMinimum}`

        if (!(keyMinimum in messages)) {
          throw new Error(
            `Zod Intl: ${keyMinimum} id is not defined in messages`
          )
        }

        message = intl.formatMessage(
          messages[keyMinimum as keyof typeof messages],
          { minimum }
        )
        break
      case 'too_big':
        const maximum =
          (issue as any).origin === 'date'
            ? dayjs.unix((issue as any).maximum as number).format('LL')
            : (issue as any).maximum
        const precisionMaximum = (issue as any).exact
          ? 'exact'
          : (issue as any).inclusive
            ? 'inclusive'
            : 'not_inclusive'
        const keyMaximum = `too_big_${(issue as any).origin}_${precisionMaximum}`

        if (!(keyMaximum in messages)) {
          throw new Error(
            `Zod Intl: ${keyMaximum} id is not defined in messages`
          )
        }

        message = intl.formatMessage(
          messages[keyMaximum as keyof typeof messages],
          { maximum }
        )
        break
      case 'custom':
        if (!(issue as any)?.params?.id) {
          throw new Error(
            `Zod Intl: Custom validation with path ${issue.path || []} has params.id undefined`
          )
        }

        if (!((issue as any).params.id in messages)) {
          throw new Error(
            `Zod Intl: ${(issue as any).params.id} id is not defined in messages`
          )
        }

        message = intl.formatMessage(
          get(messages, (issue as any).params.id),
          (issue as any).params
        )
        break
    }
    return { message }
  }
}
