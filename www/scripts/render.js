const marked = require('marked')
const mustache = require('mustache')
const fs = require('fs')
const path = require('path')
const mkdirp = require('mkdirp')
const merge = require('merge')

const postDir = path.join(process.cwd(), process.argv[2])
const outDir = path.join(process.cwd(), process.argv[3])

function expand(post) {
  const metaFile = require.resolve(post)
  const dir = path.dirname(metaFile)
  const meta = require(metaFile)
  const base = meta.extends ? expand(meta.extends) : {}
  return merge.recursive(base, meta)
}

function render(postDir, outDir) {
  const meta = expand(postDir)

  // console.log(meta)
  if (!meta.template) {
    throw new Error('no template ' + meta.pageTitle)
  }

  const template = fs.readFileSync(meta.template, 'utf8')
  const partials = Object.keys(meta.partials).reduce((partials, partialName) => {
    const partialFile = meta.partials[partialName]
    const partial = fs.readFileSync(partialFile, 'utf8')
    partials[partialName] = /.md$/.test(partialFile) ? marked(partial) : partial
    return partials
  }, {})

  const out = mustache.render(template, Object.assign({}, meta, {
    ga: process.env.GA,
  }), partials)

  const finalDir = path.join(outDir, meta.url)
  mkdirp.sync(finalDir)
  fs.writeFileSync(path.join(finalDir, 'index.html'), out, 'utf8')
  if (meta.assets) {
    meta.assets.forEach(asset => {
      fs.createReadStream(asset).
      pipe(fs.createWriteStream(path.join(finalDir, path.basename(asset))))
    })
  }
}

render(postDir, outDir)
