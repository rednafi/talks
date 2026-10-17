import { defineShikiSetup } from '@slidev/types'

// Tokyo Night sits close to atelier's cool periwinkle palette.
// Don't/Do comments get their own classes so they stand out.
export default defineShikiSetup(() => ({
  themes: {
    dark: 'tokyo-night',
    light: 'tokyo-night',
  },
  transformers: [
    {
      name: 'dont-do-comments',
      span(node, _line, _col, _lineEl, token) {
        const t = token.content.trim()
        if (/^(\/\/|#) Don't\b/.test(t)) this.addClassToHast(node, 'cm-dont')
        else if (/^(\/\/|#) Do\b/.test(t)) this.addClassToHast(node, 'cm-do')
      },
    },
  ],
}))
