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

ALTER TABLE IF EXISTS public.urls
    ADD COLUMN "authorId" bigint NOT NULL;

create index short_url on
urls (short_url);

create index original_url on
urls (original_url);

alter table public.urls add constraint urls_original_unique unique (original_url);

ALTER TABLE IF EXISTS public.urls
    ADD CONSTRAINT "FK_AUTHOR" FOREIGN KEY ("authorId")
    REFERENCES public."USERS" ("USERID") MATCH SIMPLE
    ON UPDATE NO ACTION
    ON DELETE NO ACTION
    NOT VALID;
CREATE INDEX IF NOT EXISTS "fki_FK_AUTHOR"
    ON public.urls("authorId");
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE public.urls;
-- +goose StatementEnd
