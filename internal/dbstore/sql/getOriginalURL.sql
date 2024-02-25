select
	u.original_url, u.deletedflag
from
	urls u
where
	u.short_url = $1