import { rankWith, uiTypeIs } from '@jsonforms/core'

export const collapsibleGroupTester = rankWith(3, uiTypeIs('CollapsibleGroup'))
