import { TemplateQueryKeys } from './query-keys'

describe('TemplateQueryKeys', () => {
  const orgId = 'org-123'
  const templateId = 'template-456'

  describe('list', () => {
    it('should generate list query key without filters', () => {
      const key = TemplateQueryKeys.list(orgId)
      expect(key).toEqual(['templates', 'list', orgId])
    })

    it('should generate list query key with filters', () => {
      const filters = { name: 'test', page: 1, limit: 10 }
      const key = TemplateQueryKeys.list(orgId, filters)
      expect(key).toEqual([
        'templates',
        'list',
        orgId,
        { name: 'test', page: 1, limit: 10 }
      ])
    })

    it('should filter out undefined and empty values', () => {
      const filters = {
        name: 'test',
        outputFormat: '' as any,
        createdAt: undefined,
        limit: 10
      }
      const key = TemplateQueryKeys.list(orgId, filters)
      expect(key).toEqual([
        'templates',
        'list',
        orgId,
        { name: 'test', limit: 10 }
      ])
    })

    it('should return base key when all filters are empty', () => {
      const filters = {
        name: '',
        outputFormat: '' as any,
        createdAt: undefined
      }
      const key = TemplateQueryKeys.list(orgId, filters)
      expect(key).toEqual(['templates', 'list', orgId])
    })
  })

  describe('detail', () => {
    it('should generate detail query key', () => {
      const key = TemplateQueryKeys.detail(orgId, templateId)
      expect(key).toEqual(['templates', 'detail', orgId, templateId])
    })
  })

  describe('mutations', () => {
    it('should generate create mutation key', () => {
      const key = TemplateQueryKeys.mutations.create(orgId)
      expect(key).toEqual(['templates', 'create', orgId])
    })

    it('should generate update mutation key', () => {
      const key = TemplateQueryKeys.mutations.update(orgId, templateId)
      expect(key).toEqual(['templates', 'update', orgId, templateId])
    })

    it('should generate delete mutation key', () => {
      const key = TemplateQueryKeys.mutations.delete(orgId)
      expect(key).toEqual(['templates', 'delete', orgId])
    })
  })

  describe('invalidation helpers', () => {
    it('should generate all templates key', () => {
      const key = TemplateQueryKeys.all(orgId)
      expect(key).toEqual(['templates', orgId])
    })

    it('should generate all lists key', () => {
      const key = TemplateQueryKeys.allLists(orgId)
      expect(key).toEqual(['templates', 'list', orgId])
    })
  })
})
