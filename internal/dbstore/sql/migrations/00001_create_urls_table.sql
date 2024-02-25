-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.urls
(
    "ID" bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 1000000 CACHE 1 ),
    short_url text COLLATE pg_catalog."default" NOT NULL,
    original_url text COLLATE pg_catalog."default",
    "authorId" text COLLATE pg_catalog."default" NOT NULL,
    deletedflag boolean NOT NULL DEFAULT false,
    CONSTRAINT urls_pkey PRIMARY KEY ("ID"),
    CONSTRAINT urls_original_unique UNIQUE (original_url)
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.urls
    OWNER to postgres;
-- Index: original_url

-- DROP INDEX IF EXISTS public.original_url;

CREATE INDEX IF NOT EXISTS original_url
    ON public.urls USING btree
    (original_url COLLATE pg_catalog."default" ASC NULLS LAST)
    TABLESPACE pg_default;
-- Index: short_url

-- DROP INDEX IF EXISTS public.short_url;

CREATE INDEX IF NOT EXISTS short_url
    ON public.urls USING btree
    (short_url COLLATE pg_catalog."default" ASC NULLS LAST)
    TABLESPACE pg_default;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE public.urls;
-- +goose StatementEnd
