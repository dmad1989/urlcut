-- +goose Up
-- +goose StatementBegin
create table if not exists public.urls
	(
		"ID" bigint not null generated always as identity ( increment 1 start 1 minvalue 1 maxvalue 1000000 cache 1 ),
		short_url text collate pg_catalog."default" not null,
		original_url text collate pg_catalog."default",
		constraint urls_pkey primary key ("ID")
	)	
	tablespace pg_default;

alter table if exists public.urls owner to postgres;

create index short_url on
urls (short_url);

create index original_url on
urls (original_url);

alter table public.urls add constraint urls_original_unique unique (original_url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE public.urls;
-- +goose StatementEnd
