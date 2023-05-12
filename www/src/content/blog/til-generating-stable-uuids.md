---
title: "TIL: A use case for UUIDv5"
description: "In this post I talk about generating stable IDs using UUIDv5 and my use case for them"
pubDate: "2023-05-12"
---

📓 Today I learned (well, technically it was a week ago but it took a while to write this) that you can generate stable UUIDs. When you hear "UUID" you probably think of [UUIDv4](https://datatracker.ietf.org/doc/html/rfc4122#section-4.4) which is the most popular algorithm for generating UUIDs. This one _is_ completely (pseudo)random.

There are different versions of UUIDs, and some of them are stable. Versions 3 and 5 of UUIDs are generated from "names" within a "name space". The UUID algorithm then maps names to UUIDs. If two names map to the same UUID then that means the names must be the same (sounds like the definition of a hash function, right?). The difference between V3 and V5 is how the name is hashed, V3 uses MD-5 while V5 hashes with SHA-1. The [RFC recommends](https://datatracker.ietf.org/doc/html/rfc4122#section-4.3) using V5 over V3 so that's what I'll be showcasing here.

```js
// example in Node.js using the `uuid` package
const { v5 } = require("uuid");

 // just some random UUID
let NS = "1b671a64-40d5-491e-99b0-da01ff1f3341";
v5("foo", NS); // => '71be6641-27bb-56b6-8511-571409bb0406'
// do this again and you'll get back the same UUID
v5("foo", NS); // => '71be6641-27bb-56b6-8511-571409bb0406'

// change the namespace
NS = "874c95a0-3631-4093-895e-34eb0cdafc13";
// and you get a different result
v5("foo", NS); // => '86ff592f-f174-511e-b82f-ba0a8781992b'
```

As long the namespace doesn't change you'll get the same UUID back for the same string. The namespace can be part of your code's configuration or even a constant in the source, just make sure it doesn't change 😄.

## Use case: sync data to Sanity

I needed to sync some data I had in a CSV to a [Sanity](https://www.sanity.io/) dataset. I wanted my sync script to be idempotent so that running it mutliple times, without nuking the dataset wouldn't create a bunch of duplicate documents.

My first go at it I prefixed my IDs with `csv.`, e.g. `csv.foo`, `csv.bar`. This allowed me to run "uspert" commands on the Sanity dataset:

```js
client.mutate(data.map((row) => ({
  createOrReplace: {
    _id: `csv.${row[0]}`,
    _type: "aDocument",
    // ...
  },
})))
```

However, if IDs contain a full stop then for security reasons Sanity makes those documents _invisible_ to anonymous requests.

The reason Sanity does this is due to how it keeps track of draft documents: it creates a new document prefixed with `draft.` followed by the the same ID as the original. When an anonymous request comes in, Sanity doesn't want to reveal its drafts, it returns only published documents. This is a good thing, but it meant I had to authenticate all of my requests that retrieve content from the dataset. As a bonus now I also had to make sure I didn't accidentally retrieve any draft documents. 😩

This is where UUIDv5 came in clutch here, I could generate a stable UUID for each row in the CSV and use that as the ID in Sanity:

```js
const { v5 } = require("uuid");
const NS = "aa5b0e61-b102-49a8-a878-1d6ef731469b";

client.mutate(data.map((row) => ({
  createOrReplace: {
    _id: v5(row[0], NS),
    _type: "aDocument",
    // ...
  },
}))
```

With proper UUIDs, requests to Sanity became much simpler and I didn't have to worry about accidentally retrieving draft documents. 😬
