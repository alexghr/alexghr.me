---
title: "Recovering a file from the git index"
description: ""
pubDate: "2019-08-06"
---

Ever lost a file that you knew for sure you added to git?

Well, you'll be happy to hear that git stores every file that's added to the
index and we can use that to retrieve any file git has seen. The problem is
they're all addressed by SHA1 hashes of their content.

The `.git/objects` directory contains all of the objects git has recorded. The
directory contains lots of subdirectories, which are just the two first letter
of the object hash. Inside each subdirectory you will find the actual blobs.
You can't read them directly, but you can instruct git to print the file to
stdout for you using the `git cat-file` command.

I wrote the following script to concatenate all of the git objects and print
their contents. Pipe everything through `less` and search for something that's
in the file you're looking for. There might be multiple versions of the file so
look carefully. Once you found the file, you can either copy the content to a
new file on disk or you can take the object hash that was printed above the
content and give that to `git cat-file -p <hash>` to print only that blob.

```bash
#!/usr/bin/env bash

for dir in ./.git/objects/*; do
  hashPrefix=$(basename "$dir")

  if [ "$hashPrefix" == "pack" ] || [ "$hashPrefix" == "info" ]; then
    continue
  fi


  for file in $dir/*; do
    hashSuffix=$(basename "$file")
    object="$hashPrefix$hashSuffix"

    echo "========================================"
    echo "object: $object"
    git cat-file -p "$object"
  done
done
```

Reference:

- The Pro Git Book: https://git-scm.com/book/en/v2/Git-Internals-Git-Objects
