select u.short_url, u.original_url from public.urls u where u."authorId" = $1
