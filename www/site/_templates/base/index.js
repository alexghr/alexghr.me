const moment = require('moment')

module.exports = {
  template: __dirname + '/base.mustache',
  env: {
    ga: process.env.GA,
    googleSiteVerif: process.env.GOOGLE_SITE_VERIF,
    printDate: () => (text, render) => {
      try {
        const [ dateString, format ] = render(text).split(',')
        const date = moment(dateString)
        return date.format(format)
      } catch (e) {
        console.log(e)
        return ''
      }
    }
  },
  partials: {
    ga: __dirname + '/ga.mustache'
  }
}
