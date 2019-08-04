module.exports = {
  pageTitle: 'Rubber ducking a blog',
  url: '/blog/rubber ducking a blog',
  extends: __dirname + '/../../../_templates/redirect',
  list: false,
  partials: {
    newLink: __dirname + '/link.txt'
  }
}
