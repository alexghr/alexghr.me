const getBlogPosts = require('../../util').getBlogPosts
const posts = getBlogPosts(__dirname + '/../blog')

module.exports = {
  url: '/',
  extends: __dirname + '/../../_templates/base',
  env: {
    posts
  },
  partials: {
    body: __dirname + '/index.html'
  }
}
