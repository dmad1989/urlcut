select
	u.original_url
from
	urls u
where
	u.short_url = $1