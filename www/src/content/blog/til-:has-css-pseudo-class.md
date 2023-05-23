---
title: "TIL: Use :has to target parent elements in CSS selectors"
description: "Today I learned: about the :has CSS pseudo class and how it can be used to target parent elements."
pubDate: "2023-05-23T00:00:00Z"
---

When I was first introduced to CSS in university, I was told that selectors only go from left to right. No matter how complex a selector is, the declaration will always apply to the last element in the chain.

Well, with the `:has` pseudo class achieving pretty decent browser support that's no longer true. It's probably pretty bad for performance, but for those one off cases it's useful to have in the toolkit.

The `:has` pseudo class works by selecting elements that also match the selector arguments passed to `:has`. For example:

```css
section:has(> h1) {
  color: red;
}
```

This rule will make the text color `red` for every `<section>` that has an `<h1>` element as a direct descendant. Notice that the declaration is applied to the `section` element, not the `h1` element (as would've happened with `section > h1`). You can read more about `:has` on [MDN](https://developer.mozilla.org/en-US/docs/Web/CSS/:has).

As of May 2023, `:has` is supported by Chrome (105+, since August 2022), Edge(105+, since September 2022) and Safari (15.4, since March 2022) but Firefox still has it behind a feature flag. Mobile support is pretty decent as well (but it requires much more browsers from 2023 or newer). Full table on [Can I use](https://caniuse.com/css-has).

## Use case: add overlay behind Osano GDPR banner

My use case has been to add an overlay behind a GDPR banner. The banner is added to the DOM at runtime and I didn't want to use JS for this because it would've meant continously listening for DOM mutations to detect when the banner was shown or removed. My solution in plain CSS:

```css
.osano-cm-window:has(> .osano-cm-dialog)::before {
  content: "";
  position: fixed;
  inset: 0;
  background-color: rgb(0 0 0 / 40%);
}

.osano-cm-window:has(> .osano-cm-dialog--hidden)::before {
  content: none;
}
```

`.osano-cm-window` is the root element that Osano adds to the DOM. Everything that their script adds, lives inside of this element. `.osano-cm-dialog` is the element that contains the banner and `.osano-cm-dialog--hidden` is added to the banner when it's closed.

My biggest worry with this was the overlay breaking the page if `:has` wasn't supported by the browser, but by using `:has` in both declarations it gracefully degrades and the overlay isn't added to the DOM at all if `:has` is not available.
