# retryhttp

Provides RetryRoundTripper used by Baggageclaim client and ATC garden client.

Retries on network errors, does not retry if request body was already read from (e.g. streaming request)
