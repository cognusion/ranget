# ranget

ranget (pronounced rang-et or range-et (I say the latter) or ran-jay if you're Frenchish, I suppose) is a CLI tool that fetches the ``--url``-specified file over HTTP/S using the HTTP RANGE spec to download ``--size`` chunks of the file asynchonously using the ``--max`` number of workers. When not ``--debug``ging, there is a nice progress bar that also calculates the throughput of the workers. If the requested server does not support ranged requests, the file is downloaded "normally".

ranget is an example driver for [rangetripper](https://github.com/cognusion/go-rangetripper/v2) which you can use directly.

## Usage

```quote
      --debug              Enable debugging output (disables progress bar)
      --max int            Maximum number of simultaneous downloaders (default 10)
      --out string         Where to write it it (default "./afile")
      --size string        Size of chunks to download (whole-numbers with suffixes of B,KB,MB,GB,PB) (default "10MB")
      --timeout duration   Set a general timeout for the download (default 1m0s)
      --trash              Delete the file after downloading (e.g. if benchmarking, etc.)
      --url string         What to fetch
```
