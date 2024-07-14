Kobo Browser Quirks

Apparently the browser is using a very old version of WebKit.

- Doesn't support `fetch`
    - No HTMX
- Doesn't support `secure` or `httpOnly` cookies
    - They just silently fail to be set with these flags
- Makes 2 parallel requests whenever an `<a>` link is clicked
    - This doesn't seem to apply to URL bar navigation
    - This poses issues when a request is not idempotent. Need to figure out a solution
      for these cases... For example when converting to Kepub, we encounter failures
      for the second requests due to file conflicts / deletions happening at the same time.
      I've fixed this by locking the conversion to a single request at a time with a mutex,
      but we still do the conversion twice, just one after the other.
      I was planning on creating an OPDS interface for OpenBooks but this will make all
      search / download requests send twice which is no good. The fix isn't as simple
      in that case.