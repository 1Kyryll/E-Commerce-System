### 6. Catalog reads - cache-aside and cursor pagination 

Request for a catalog hits the Redis cache first(name, description, image, stock is left for DB query), if nothing is found, then hit the DB directly. Use cursor pagination for queryinh chunks of data.

**Tradeoff:** if we also want *in-stock* property in cache, the data might be stale for a couple of seconds. However, it isn't a big deal because if the user clicks such an item with slightly stale data, he will immediately get rejected. Cursor pagination might introduce complications related to cursor(uuid, published_at) and DB's data type casting. 

**Alternatives:**
- *Offset pagination.* It is more straightforward than cursor pagination, but that comes with probability of page loss/appearance under constant scrolls or adding/deleting products to DB.
- *No cache.* Every page render hits the DB, which doesn't scale quite well.
- *Cache inventory count.* Inventory count must be consistent, it is constantly changing, what doesn't overlap with caching concept.
