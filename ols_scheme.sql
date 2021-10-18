--
-- PostgreSQL database dump
--

-- Dumped from database version 12.1
-- Dumped by pg_dump version 12.1

--
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: EXTENSION pg_trgm; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION pg_trgm IS 'text similarity measurement and index searching based on trigrams';


SET default_tablespace = '';

SET default_table_access_method = heap;


--
-- Name: farma; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.farma (
    id integer NOT NULL,
    name character varying(200) NOT NULL,
    is_znvlp integer DEFAULT 0 NOT NULL
);


--
-- Name: farma_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.farma_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: farma_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.farma_id_seq OWNED BY public.farma.id;


--
-- Name: person; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.person (
    id integer,
    fio character varying(150) NOT NULL,
    ndoc character varying(10) NOT NULL,
    active integer,
	dossier_num varchar(20) default ''::character varying not null,
	postal_code varchar(10) default ''::character varying not null,
	address text default ''::text not null
);


--
-- Name: prp; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.prp (
    id integer,
    id_own integer,
    num character varying(30),
    dtbeg date,
    dtend date,
    active integer,
    is_med integer DEFAULT 0
);


--
-- Name: prp_template; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.prp_template (
    id_own integer,
    id integer NOT NULL,
    id_farm integer DEFAULT 0 NOT NULL,
    cnt integer DEFAULT 0 NOT NULL,
    name character varying(100),
    mark_del integer DEFAULT 0 NOT NULL
);


--
-- Name: prp_template_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.prp_template_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: prp_template_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.prp_template_id_seq OWNED BY public.prp_template.id;


--
-- Name: visit; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.visit (
    id integer NOT NULL,
    dt date,
    id_prp integer,
    id_own integer
);


--
-- Name: visit_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.visit_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: visit_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.visit_id_seq OWNED BY public.visit.id;


--
-- Name: visit_info; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.visit_info (
    id_own integer,
    id_prpt integer,
    cnt integer DEFAULT 0 NOT NULL,
    price numeric(10,2) DEFAULT 0 NOT NULL,
    price_znvlp numeric(10,2) DEFAULT 0 NOT NULL,
    reason character varying(100) DEFAULT ''::character varying NOT NULL,
    paydt character varying(100) DEFAULT ''::character varying NOT NULL,
    prevcnt integer DEFAULT 0 NOT NULL,
    cnt_recep integer DEFAULT 0 NOT NULL
);


--
-- Name: farma id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.farma ALTER COLUMN id SET DEFAULT nextval('public.farma_id_seq'::regclass);


--
-- Name: prp_template id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prp_template ALTER COLUMN id SET DEFAULT nextval('public.prp_template_id_seq'::regclass);


--
-- Name: visit id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.visit ALTER COLUMN id SET DEFAULT nextval('public.visit_id_seq'::regclass);


--
-- Name: farma farma_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.farma
    ADD CONSTRAINT farma_pkey PRIMARY KEY (id);


--
-- Name: prp_template prp_template_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.prp_template
    ADD CONSTRAINT prp_template_pkey PRIMARY KEY (id);


--
-- Name: visit visit_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.visit
    ADD CONSTRAINT visit_pkey PRIMARY KEY (id);


--
-- Name: aPERSON; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "aPERSON" ON public.person USING btree (active);


--
-- Name: aPRP; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "aPRP" ON public.prp USING btree (active);


--
-- Name: idFARMA; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idFARMA" ON public.farma USING btree (id);


--
-- Name: idPERSON; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idPERSON" ON public.person USING btree (id);


--
-- Name: idPRP; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idPRP" ON public.prp USING btree (id);


--
-- Name: idPRPT; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idPRPT" ON public.prp_template USING btree (id);


--
-- Name: idVISIT; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idVISIT" ON public.visit USING btree (id);


--
-- Name: idownPRP; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idownPRP" ON public.prp USING btree (id_own);


--
-- Name: idownPRPT; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idownPRPT" ON public.prp_template USING btree (id_own);


--
-- Name: idownVI; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idownVI" ON public.visit_info USING btree (id_own);


--
-- Name: idownVISIT; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idownVISIT" ON public.visit USING btree (id_own);


--
-- Name: idprpVISIT; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idprpVISIT" ON public.visit USING btree (id_prp);


--
-- Name: idprptVI; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "idprptVI" ON public.visit_info USING btree (id_prpt);


--
-- Name: nameFARMA; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX "nameFARMA" ON public.farma USING gist (name public.gist_trgm_ops);


--
-- PostgreSQL database dump complete
--

