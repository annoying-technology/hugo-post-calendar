# Hugo Post Calendar

This is a small utility that is part of the [Annoying.Technology](https://annoying.technology) build pipeline.

It generates a iCalendar file based on the “Published At” timestamp of Hugo posts. This helps us to schedule posts better so we don’t post twice per day. This file needs to be hosted somewhere so Calendar.app can subscribe to it.

## Usage

```
go build && ./hugo-post-calendar --source="/Users/philipp/Blog/annoying.technology" --destination="annoying-technology.ics"
```

Or just use the binary on the [Releases](https://github.com/annoying-technology/hugo-post-calendar/releases) page.