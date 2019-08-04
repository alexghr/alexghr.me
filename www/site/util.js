const fs = require('fs')
const path = require('path')

module.exports.getBlogPosts = function(blogDir) {
  const postDirs = fs.readdirSync(blogDir)
  const posts = postDirs
    .map(postDir => require(path.join(blogDir, postDir)))
    .filter(post => post.list)
    .sort((postA, postB) => {
      const dateA = Date.parse(postA.date)
      const dateB = Date.parse(postB.date)

      return dateB - dateA
    })

  return posts
}
