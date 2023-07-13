---
title: "The effects of the HN 'personal blogs' thread on my RSS feed"
description: "A recent trend in the HN community has been to share personal websites. In this article I look at the RSS traffic I got after posting this site."
pubDate: "2023-07-13"
---

A recent thread on Hacker News created quite a bit of buzz in the community. [Ask HN: Could you share your personal blog here?](https://news.ycombinator.com/item?id=36575081) got almost 2000 comments of people posting about their blogs. I missed this thread entirely but a couple of days later a related post made it to the HN homepage.

[Show HN: OPML list of Hacker News Users Personal Blogs](https://news.ycombinator.com/item?id=36627112) took all of the blogs from the original thread that had an RSS feed and put them in a list for easy importing into RSS feed readers. That's is when [I added my blog to the original Ask HN thread](https://news.ycombinator.com/item?id=36627962), on 2023-07-07, 4 days after the original Ask HN thread was started. Fortunately the author of the OPML list re-ran the import script and my website's RSS feed was picked up.

A couple more side projects using the blogs from the original Ask HN blog popped up and made it to the front page. I got curious if all this attention had generated any meaningful traffic to my RSS feed.

## Setup

I migrated my website to a new VPS last month so I've only got logs since 2023-06-11. I will be using standard UNIX shell tools to analyse the logs and gnuplot to plot the data. I'll be publishing the aggregated log data and the scripts I used to extract it.

## Requests for `rss.xml`

I'll start with a plot of requests per day for this site's feed:

<figure style="display:block;margin:2em auto;">

![plot](/blog/effects-of-hn-personal-blogs-thread/rss-hits-per-day.svg)
  <figcaption style="text-align:center">Daily requests for <code>/rss.xml</code></figcation>
</figure>

Basically a flat graph until 2023-07-07 when I first posted the website on the Ask HN thread. On July 7th it goes up to almost 300 requests, on July 9th to 1.2k requests/day and looks like it's slowly been growing since. There is a small bump on 2023-06-20 of _2_ requests. That was just me testing the feed. The graph looks like it's fallen of a cliff at the end because the last day is incomplete 😬

You can grab a copy of the data here: [rss-hits-per-day.csv](/blog/effects-of-hn-personal-blogs-thread/rss-hits-per-day.csv).

<details>
<summary>I extracted the data using this script</summary>

```bash
# get only requests to /rss.xml
$ grep /rss.xml access.log* \
# take the time of each request, 4th field as read by awk
  | awk '{ print $4 }' \
# nginx prints dates between square brackets, remove them
# for a period of time I didn't use ISO8601 timestamps, fix them
  | sed -e 's/\[//g' -e 's/\]//g' -e 's/\// /g' -e 's/2023:/2023 /' \
# parse datetime and only keep the date
  | xargs -I D date --date D --iso-8601=date \
  | sort \
  | uniq -c \
  | awk '{print $2","$1}' \
  > rss-hits-per-day.csv
```

Then I manually filled the missing dates since 2023-06-11 with 0s.
</details>

<details>
<summary>This is the gnuplot script for the chart above</summary>

```gnuplot
set datafile separator ','

set xdata time
set timefmt "%Y-%m-%d"
set format x "%Y-%m-%d"

set xlabel "Date"
set ylabel "Hits to /rss.xml"

set terminal svg size 800,600 font "Helvetica,14" background "#ffffff"

set output "rss-hits-per-day.svg"
set key autotitle columnhead

plot "rss-hits-per-day.csv" using 1:2 with lines, '' using 1:3 with lines
```

</details>

## User agents

The other thing I was interested in was what user agents were being used to access the feed. The top 5 most frequent user agent strings are:

<table border="0" style="margin:2em auto;min-width:60%">
  <thead style="font-weight:700">
    <tr>
      <td>RSS reader</td>
      <td>Count</td>
    </tr>
  </thead>
  <tbody>
    <tr><td>CommaFeed</td><td>1215</td></tr>
    <tr><td>FreshRSS</td><td>1001</td></tr>
    <tr><td>Unread RSS Reader</td><td>505</td></tr>
    <tr><td>NetNewsWire</td><td>502</td></tr>
    <tr><td>SpaceCowboys Android RSS Reader</td><td>394</td></tr>
  </tbody>
</table>

[CommaFeed](https://github.com/Athou/commafeed) seems to be a self-hosted reader so either lots of people are hosting their own version or using a third-party hosting service.

What's interesting is that some of these readers report back the number of subscribers in the user agent string. Here are some examples:

- Feedly/1.0 (+http://www.feedly.com/fetcher.html; **9 subscribers**; )
- Feedbin feed-id:2690232 - **4 subscribers**
- Mozilla/5.0 (compatible; inoreader.com; **4 subscribers**)

This muddles the data a bit because now I've got multiple entries for each, as the subscriber count increases, but it's a useful metric for content authors.

You can grab a copy of the raw user agent strings here: [rss-ua.txt](/blog/effects-of-hn-personal-blogs-thread/rss-ua.txt).

<details>
<summary>This is the script I used to extract the user agent strings from the logs</summary>

```bash
# get only requests to /rss.xml
$ grep /rss.xml access.log* \
# the user agent is the last thing printed in my logs
# awk splits it over multiple fields, join them back together
  | awk '{ out = ""; for(i = 13; i <= NF; i++) {out = out " " $i;} print out }' \
  | sort \
  | uniq -c \
  | sort -rn \
  > rss-ua.txt
```

</details>

## Conclusion

I only looked at direct traffic to the [RSS feed](/rss.xml) and didn't take into account any traffic from feed readers to actual articles. My feed only contains titles and links (no article body) so presumbly subscribers have to get to the content.

I'm surprised a small website like mine started to get this much traffic over RSS and I'm curious to see what bigger and more active blogs are getting since the HN thread.
