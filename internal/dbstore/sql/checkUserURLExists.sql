select
	exists(
	select
		1
	from
		public.urls u
	where
		u.short_url = $1
		and u."authorId" = $2)