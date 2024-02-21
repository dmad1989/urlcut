select
	u.short_url
from
	urls u
where
	u.original_url = $1