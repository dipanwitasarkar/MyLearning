# How Spotify Serves Point Queries from an Exabyte Data Lake

Scaling Reads

## The TLDR

Spotify (like all big companies) keeps most of its data in cheap object storage built to scan millions of rows at once. For decades companies have realized that data is valuable, so they almost never want to throw it away. This leads to a pattern of dumping data into inexpensive storage (a "data lake") where it's always accessible... but maybe not in the way that you want.

Data lakes are great for a weekly report or a training run. They're terrible at a single question like "what did this person listen to?" That's exactly the kind of question an online feature, or an AI agent, needs answered before the page finishes loading. The usual fix is copying whatever needs to be fast into a database built for lookups or a cache like Redis. This is so common that if you've worked at a mature company, it probably has a few clicks or a config that does just this. But databases and caches cost real money per gigabyte, so teams end up rationing what's worth serving online.

Spotify built a way around that trade they call RAP, short for Random Access Parquet which is an index living outside the data files that records exactly which file and rows hold a given key's data. This makes it so a lookup that used to mean hunting through files one clue at a time becomes a single precise read. They paired it with changes to how new files get written, so one user's data sits together instead of being scattered across many files. The innovation is that none of this needs a second copy of the data or a rewritten pipeline. The same files BigQuery scans for a weekly report also answer one person's question, at roughly the cost of a single storage read.

This is a great example of a real third option between scanning the whole lake and paying to copy everything into a key-value store.

## The Problem

### Where companies actually keep their data

Every song a Spotify user plays lands as an event somewhere. The working copy everyone uses downstream lives in what's called a data lake. That just means files sitting in object storage like Google Cloud Storage or S3, not rows in a database. Think of it like a big, cheap archive.

An analyst might query those files through a query engine like Trino or BigQuery, which reads the files where they sit rather than loading them into a database first. A machine learning pipeline might read the same files for training sets. Neither is touching a database in the traditional sense. These are big sweeping scans, not the lookups and point queries an RDBMS like Postgres is built for.

There's a ton of benefits here! Object storage is cheap enough per gigabyte that Spotify keeps exabytes in GCS, the full history rather than a recent window, since nobody has to ration what gets kept. And every tool reading the same files also means one copy of the truth instead of several scattered across teams.

But, there are tradeoffs.

### Organized for scans, not lookups

Picture an ordinary question: "what is the average listening time by country over the last week?" Logically, this means touching two columns (country and listening time) across every row in that window. Spotify's lake files are Parquet files and they're made for queries just like this. Parquet is organized column by column rather than row by row. Every value from one column is packed together, then the next. The file gets sliced into row groups, each row group holds a chunk of every column, those chunks get sliced again into pages, and a footer describes where everything lives.

Values from the same column compress well because they're similar. An engine can also skip whole pages it doesn't need without reading them.

To answer our question, we just have to grab these two columns from every file in the window and do some math. It's a lot of data to scan, but it's exactly the data that we need. Data lakes work pretty well for this kind of query!

Now flip the question and ask what one specific person listened to. That's a handful of rows scattered across whatever files hold that user's data, and the layout that made our average-listening-time query fast does nothing for this one. If you can't pin down exactly where that row lives, you end up dragging gigantic pages of data off storage to get at a few hundred bytes. On top of that you pay for a full query plan, because Trino and BigQuery schedule every query as a distributed analytical job, even one that wants a single row.

That gap is why online features copy their data out. Whatever needs to be fast goes into Bigtable or DynamoDB, a second copy priced per gigabyte. If Spotify is like other companies, they'll either have budgets or goals to reduce the size of this tier. So everything else stays in the lake, still around but too slow to serve a page from.

## The Solution

### Narrowing ninety thousand files down to twelve

Start with the scale of the problem. Across Spotify's hundreds of millions of users, a single day's listening events land in about a thousand files. So take that same question about one person's listening, the one we just asked, and put a window on it: a roughly 90-day summer, asked either by the user opening their profile or by an agent answering on their behalf. At 1,000 files a day, that's 90,000 candidate Parquet files. This is gigabytes to terabytes of data to scan which is just not practical in any online setting.

What many teams do is organize their data along the dimensions their queries actually use. If lookups by user are common, then within each day the files get bucketed by user ID: hash the ID, mod it by the number of files, and that's where that user's events go. This narrows the search space dramatically, but you still need to scan within those files to find the specific rows.

### RAP: Random Access Parquet

Spotify's innovation is RAP, an index that lives outside the Parquet files. The index records exactly which file and which rows contain a given key's data. Instead of scanning files to find a user's data, you query the index to get the precise location, then read exactly those rows.

The index is built as new files are written, so it stays in sync with the data. When a query comes in for a specific user, the index returns the exact file paths and row offsets, and you can read just those rows with a single precise operation.

This means the same files that BigQuery scans for weekly reports can also answer point queries, without needing a second copy of the data in a separate database.

## Key Takeaways

1. **Data lakes are great for scans, terrible for lookups**: Object storage is cheap and great for analytical queries, but terrible for point queries.

2. **Indexes can bridge the gap**: An external index that maps keys to precise file locations can make data lakes usable for point queries.

3. **Avoid data duplication**: With the right indexing, you can serve both analytical and operational queries from the same data, avoiding expensive duplication.

4. **AI agents increase point query demand**: As AI agents start asking more questions of data, the ability to serve point queries from data lakes becomes increasingly important.