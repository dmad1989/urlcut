select
	exists (
	select
	from 
			information_schema.tables
	where 
			table_schema like 'public'
		and 
			table_type like 'BASE TABLE'
		and
			table_name = 'urls'
		)