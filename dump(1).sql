--
-- PostgreSQL database dump
--

\restrict wXuNeXnb6IN23NbXw7myZdFuQSwJmMpFTo7lgewyepthXjaaNX25DZDBeABLA3b

-- Dumped from database version 17.7 (Ubuntu 17.7-0ubuntu0.25.04.1)
-- Dumped by pg_dump version 17.7 (Ubuntu 17.7-0ubuntu0.25.04.1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

ALTER TABLE IF EXISTS ONLY public.user_subscriptions DROP CONSTRAINT IF EXISTS user_subscriptions_user_id_fkey;
ALTER TABLE IF EXISTS ONLY public.user_subscriptions DROP CONSTRAINT IF EXISTS user_subscriptions_subscription_id_fkey;
ALTER TABLE IF EXISTS ONLY public.user_subscriptions DROP CONSTRAINT IF EXISTS user_subscriptions_payment_id_fkey;
ALTER TABLE IF EXISTS ONLY public.user_purchases DROP CONSTRAINT IF EXISTS user_purchases_user_id_fkey;
ALTER TABLE IF EXISTS ONLY public.user_purchases DROP CONSTRAINT IF EXISTS user_purchases_product_id_fkey;
ALTER TABLE IF EXISTS ONLY public.user_purchases DROP CONSTRAINT IF EXISTS user_purchases_payment_id_fkey;
ALTER TABLE IF EXISTS ONLY public.transactions DROP CONSTRAINT IF EXISTS transactions_user_id_fkey;
ALTER TABLE IF EXISTS ONLY public.subscriptions DROP CONSTRAINT IF EXISTS subscriptions_user_id_fkey;
ALTER TABLE IF EXISTS ONLY public.subscriptions DROP CONSTRAINT IF EXISTS subscriptions_plan_id_fkey;
ALTER TABLE IF EXISTS ONLY public.project_logs DROP CONSTRAINT IF EXISTS project_logs_subscription_id_fkey;
ALTER TABLE IF EXISTS ONLY public.payments DROP CONSTRAINT IF EXISTS payments_user_id_fkey;
ALTER TABLE IF EXISTS ONLY public.payments DROP CONSTRAINT IF EXISTS payments_subscription_id_fkey;
ALTER TABLE IF EXISTS ONLY public.library DROP CONSTRAINT IF EXISTS library_user_id_fkey;
ALTER TABLE IF EXISTS ONLY public.library DROP CONSTRAINT IF EXISTS library_product_id_fkey;
ALTER TABLE IF EXISTS ONLY public.library DROP CONSTRAINT IF EXISTS library_payment_id_fkey;
ALTER TABLE IF EXISTS ONLY public.files DROP CONSTRAINT IF EXISTS files_created_by_fkey;
ALTER TABLE IF EXISTS ONLY public.discounts DROP CONSTRAINT IF EXISTS discounts_created_by_user_id_fkey;
DROP TRIGGER IF EXISTS update_users_updated_at ON public.users;
DROP TRIGGER IF EXISTS update_subscriptions_updated_at ON public.subscriptions;
DROP TRIGGER IF EXISTS update_products_updated_at ON public.products;
DROP INDEX IF EXISTS public.idx_users_username;
DROP INDEX IF EXISTS public.idx_users_referral_code;
DROP INDEX IF EXISTS public.idx_users_email;
DROP INDEX IF EXISTS public.idx_users_created_at;
DROP INDEX IF EXISTS public.idx_users_balance;
DROP INDEX IF EXISTS public.idx_user_subscriptions_user_id;
DROP INDEX IF EXISTS public.idx_user_subscriptions_expires_at;
DROP INDEX IF EXISTS public.idx_user_purchases_user;
DROP INDEX IF EXISTS public.idx_user_purchases_product;
DROP INDEX IF EXISTS public.idx_user_purchases_date;
DROP INDEX IF EXISTS public.idx_transactions_user_id;
DROP INDEX IF EXISTS public.idx_subscriptions_user;
DROP INDEX IF EXISTS public.idx_subscriptions_status;
DROP INDEX IF EXISTS public.idx_subscriptions_plan_id;
DROP INDEX IF EXISTS public.idx_subscriptions_billing_date;
DROP INDEX IF EXISTS public.idx_project_logs_subscription_id;
DROP INDEX IF EXISTS public.idx_products_type;
DROP INDEX IF EXISTS public.idx_products_price;
DROP INDEX IF EXISTS public.idx_products_active;
DROP INDEX IF EXISTS public.idx_payments_user;
DROP INDEX IF EXISTS public.idx_payments_subscription;
DROP INDEX IF EXISTS public.idx_payments_status;
DROP INDEX IF EXISTS public.idx_payments_gateway_id;
DROP INDEX IF EXISTS public.idx_payments_created_at;
DROP INDEX IF EXISTS public.idx_library_user;
DROP INDEX IF EXISTS public.idx_library_product;
DROP INDEX IF EXISTS public.idx_files_is_deleted;
DROP INDEX IF EXISTS public.idx_files_created_by_deleted;
DROP INDEX IF EXISTS public.idx_files_created_by;
DROP INDEX IF EXISTS public.idx_discounts_creator;
DROP INDEX IF EXISTS public.idx_discounts_code;
DROP INDEX IF EXISTS public.idx_discounts_active;
ALTER TABLE IF EXISTS ONLY public.users DROP CONSTRAINT IF EXISTS users_username_key;
ALTER TABLE IF EXISTS ONLY public.users DROP CONSTRAINT IF EXISTS users_referral_code_key;
ALTER TABLE IF EXISTS ONLY public.users DROP CONSTRAINT IF EXISTS users_pkey;
ALTER TABLE IF EXISTS ONLY public.users DROP CONSTRAINT IF EXISTS users_email_key;
ALTER TABLE IF EXISTS ONLY public.user_subscriptions DROP CONSTRAINT IF EXISTS user_subscriptions_user_id_subscription_id_key;
ALTER TABLE IF EXISTS ONLY public.user_subscriptions DROP CONSTRAINT IF EXISTS user_subscriptions_pkey;
ALTER TABLE IF EXISTS ONLY public.user_purchases DROP CONSTRAINT IF EXISTS user_purchases_user_id_product_id_key;
ALTER TABLE IF EXISTS ONLY public.user_purchases DROP CONSTRAINT IF EXISTS user_purchases_pkey;
ALTER TABLE IF EXISTS ONLY public.library DROP CONSTRAINT IF EXISTS unique_user_product;
ALTER TABLE IF EXISTS ONLY public.transactions DROP CONSTRAINT IF EXISTS transactions_pkey;
ALTER TABLE IF EXISTS ONLY public.test DROP CONSTRAINT IF EXISTS test_pkey;
ALTER TABLE IF EXISTS ONLY public.subscriptions DROP CONSTRAINT IF EXISTS subscriptions_pkey;
ALTER TABLE IF EXISTS ONLY public.project_logs DROP CONSTRAINT IF EXISTS project_logs_pkey;
ALTER TABLE IF EXISTS ONLY public.products DROP CONSTRAINT IF EXISTS products_pkey;
ALTER TABLE IF EXISTS ONLY public.plans DROP CONSTRAINT IF EXISTS plans_pkey;
ALTER TABLE IF EXISTS ONLY public.payments DROP CONSTRAINT IF EXISTS payments_pkey;
ALTER TABLE IF EXISTS ONLY public.library DROP CONSTRAINT IF EXISTS library_pkey;
ALTER TABLE IF EXISTS ONLY public.files DROP CONSTRAINT IF EXISTS files_pkey;
ALTER TABLE IF EXISTS ONLY public.discounts DROP CONSTRAINT IF EXISTS discounts_pkey;
ALTER TABLE IF EXISTS ONLY public.discounts DROP CONSTRAINT IF EXISTS discounts_code_key;
ALTER TABLE IF EXISTS ONLY public.admin_analytics_snapshot DROP CONSTRAINT IF EXISTS admin_analytics_snapshot_pkey;
ALTER TABLE IF EXISTS public.test ALTER COLUMN id DROP DEFAULT;
ALTER TABLE IF EXISTS public.admin_analytics_snapshot ALTER COLUMN id DROP DEFAULT;
DROP TABLE IF EXISTS public.users;
DROP TABLE IF EXISTS public.user_subscriptions;
DROP TABLE IF EXISTS public.user_purchases;
DROP TABLE IF EXISTS public.transactions;
DROP SEQUENCE IF EXISTS public.test_id_seq;
DROP TABLE IF EXISTS public.test;
DROP TABLE IF EXISTS public.subscriptions;
DROP TABLE IF EXISTS public.project_logs;
DROP TABLE IF EXISTS public.products;
DROP TABLE IF EXISTS public.plans;
DROP TABLE IF EXISTS public.payments;
DROP TABLE IF EXISTS public.library;
DROP TABLE IF EXISTS public.files;
DROP TABLE IF EXISTS public.discounts;
DROP SEQUENCE IF EXISTS public.admin_analytics_snapshot_id_seq;
DROP TABLE IF EXISTS public.admin_analytics_snapshot;
DROP FUNCTION IF EXISTS public.update_updated_at_column();
DROP TYPE IF EXISTS public.subscription_status;
DROP TYPE IF EXISTS public.product_type;
DROP TYPE IF EXISTS public.payment_status;
DROP TYPE IF EXISTS public.discount_type;
DROP EXTENSION IF EXISTS "uuid-ossp";
DROP EXTENSION IF EXISTS pgcrypto;
--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- Name: discount_type; Type: TYPE; Schema: public; Owner: pixelcraft_user
--

CREATE TYPE public.discount_type AS ENUM (
    'PERCENTAGE',
    'FIXED_AMOUNT'
);


ALTER TYPE public.discount_type OWNER TO pixelcraft_user;

--
-- Name: payment_status; Type: TYPE; Schema: public; Owner: pixelcraft_user
--

CREATE TYPE public.payment_status AS ENUM (
    'PENDING',
    'COMPLETED',
    'FAILED',
    'REFUNDED'
);


ALTER TYPE public.payment_status OWNER TO pixelcraft_user;

--
-- Name: product_type; Type: TYPE; Schema: public; Owner: pixelcraft_user
--

CREATE TYPE public.product_type AS ENUM (
    'PLUGIN',
    'MOD',
    'MAP',
    'TEXTUREPACK',
    'SERVER_TEMPLATE'
);


ALTER TYPE public.product_type OWNER TO pixelcraft_user;

--
-- Name: subscription_status; Type: TYPE; Schema: public; Owner: pixelcraft_user
--

CREATE TYPE public.subscription_status AS ENUM (
    'ACTIVE',
    'CANCELED',
    'PAST_DUE'
);


ALTER TYPE public.subscription_status OWNER TO pixelcraft_user;

--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: pixelcraft_user
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_updated_at_column() OWNER TO pixelcraft_user;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: admin_analytics_snapshot; Type: TABLE; Schema: public; Owner: pixelcraft_user
--

CREATE TABLE public.admin_analytics_snapshot (
    id integer NOT NULL,
    total_revenue numeric(10,2) DEFAULT 0 NOT NULL,
    total_users integer DEFAULT 0 NOT NULL,
    active_products integer DEFAULT 0 NOT NULL,
    total_sales integer DEFAULT 0 NOT NULL,
    last_updated timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    revenue_growth_pct numeric(5,2) DEFAULT 0,
    users_growth_pct numeric(5,2) DEFAULT 0,
    sales_growth_pct numeric(5,2) DEFAULT 0,
    products_status jsonb
);


ALTER TABLE public.admin_analytics_snapshot OWNER TO pixelcraft_user;

--
-- Name: admin_analytics_snapshot_id_seq; Type: SEQUENCE; Schema: public; Owner: pixelcraft_user
--

CREATE SEQUENCE public.admin_analytics_snapshot_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.admin_analytics_snapshot_id_seq OWNER TO pixelcraft_user;

--
-- Name: admin_analytics_snapshot_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: pixelcraft_user
--

ALTER SEQUENCE public.admin_analytics_snapshot_id_seq OWNED BY public.admin_analytics_snapshot.id;


--
-- Name: discounts; Type: TABLE; Schema: public; Owner: pixelcraft_user
--

CREATE TABLE public.discounts (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    code character varying(50) NOT NULL,
    type public.discount_type NOT NULL,
    value numeric(10,2) NOT NULL,
    is_referral boolean DEFAULT false,
    created_by_user_id uuid,
    expires_at timestamp with time zone,
    max_uses integer,
    current_uses integer DEFAULT 0,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT discounts_current_uses_check CHECK ((current_uses >= 0)),
    CONSTRAINT discounts_max_uses_check CHECK ((max_uses > 0)),
    CONSTRAINT discounts_value_check CHECK ((value >= (0)::numeric))
);


ALTER TABLE public.discounts OWNER TO pixelcraft_user;

--
-- Name: TABLE discounts; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON TABLE public.discounts IS 'Cupons de desconto e cÃ³digos de indicaÃ§Ã£o';


--
-- Name: COLUMN discounts.is_referral; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.discounts.is_referral IS 'TRUE se for cÃ³digo de indicaÃ§Ã£o';


--
-- Name: COLUMN discounts.created_by_user_id; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.discounts.created_by_user_id IS 'UsuÃ¡rio que gerou o cÃ³digo de referral';


--
-- Name: files; Type: TABLE; Schema: public; Owner: pixelcraft_user
--

CREATE TABLE public.files (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(255) NOT NULL,
    file_name character varying(255) NOT NULL,
    file_type character varying(20),
    file_path character varying(500),
    size bigint,
    created_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    is_deleted boolean DEFAULT false
);


ALTER TABLE public.files OWNER TO pixelcraft_user;

--
-- Name: library; Type: TABLE; Schema: public; Owner: pixelcraft_user
--

CREATE TABLE public.library (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    product_id uuid NOT NULL,
    payment_id uuid NOT NULL,
    purchased_at timestamp with time zone DEFAULT now()
);


ALTER TABLE public.library OWNER TO pixelcraft_user;

--
-- Name: payments; Type: TABLE; Schema: public; Owner: pixelcraft_user
--

CREATE TABLE public.payments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    subscription_id uuid,
    description text NOT NULL,
    amount numeric(10,2) NOT NULL,
    discount_applied numeric(10,2) DEFAULT 0,
    final_amount numeric(10,2) NOT NULL,
    status public.payment_status DEFAULT 'PENDING'::public.payment_status NOT NULL,
    payment_gateway_id character varying(255),
    payment_method character varying(50),
    payment_metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    completed_at timestamp with time zone,
    failed_at timestamp with time zone,
    CONSTRAINT payments_amount_check CHECK ((amount >= (0)::numeric)),
    CONSTRAINT payments_discount_applied_check CHECK ((discount_applied >= (0)::numeric)),
    CONSTRAINT payments_final_amount_check CHECK ((final_amount >= (0)::numeric))
);


ALTER TABLE public.payments OWNER TO pixelcraft_user;

--
-- Name: TABLE payments; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON TABLE public.payments IS 'Registro completo de todas as transaÃ§Ãµes financeiras';


--
-- Name: COLUMN payments.subscription_id; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.payments.subscription_id IS 'NULL para compras avulsas, UUID para renovaÃ§Ãµes de assinatura';


--
-- Name: COLUMN payments.payment_gateway_id; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.payments.payment_gateway_id IS 'ID externo do gateway de pagamento';


--
-- Name: COLUMN payments.payment_method; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.payments.payment_method IS 'MÃ©todo usado: BALANCE, STRIPE, MERCADOPAGO, etc.';


--
-- Name: plans; Type: TABLE; Schema: public; Owner: pixelcraft_user
--

CREATE TABLE public.plans (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(100) NOT NULL,
    description text,
    price numeric(10,2) NOT NULL,
    is_active boolean DEFAULT true,
    features jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT plans_price_check CHECK ((price >= (0)::numeric))
);


ALTER TABLE public.plans OWNER TO pixelcraft_user;

--
-- Name: products; Type: TABLE; Schema: public; Owner: pixelcraft_user
--

CREATE TABLE public.products (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    price numeric(10,2) NOT NULL,
    type public.product_type NOT NULL,
    download_url_encrypted bytea NOT NULL,
    is_exclusive boolean DEFAULT false,
    stock_quantity integer,
    image_url text,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT products_price_check CHECK ((price >= (0)::numeric)),
    CONSTRAINT products_stock_quantity_check CHECK ((stock_quantity >= 0))
);


ALTER TABLE public.products OWNER TO pixelcraft_user;

--
-- Name: TABLE products; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON TABLE public.products IS 'CatÃ¡logo de produtos digitais (plugins, mods, maps, etc.)';


--
-- Name: COLUMN products.download_url_encrypted; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.products.download_url_encrypted IS 'URL de download criptografada com pgp_sym_encrypt';


--
-- Name: COLUMN products.is_exclusive; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.products.is_exclusive IS 'Produto de ediÃ§Ã£o limitada (venda Ãºnica)';


--
-- Name: COLUMN products.stock_quantity; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.products.stock_quantity IS 'Quantidade em estoque (NULL = ilimitado)';


--
-- Name: project_logs; Type: TABLE; Schema: public; Owner: pixelcraft_user
--

CREATE TABLE public.project_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    subscription_id uuid NOT NULL,
    message text NOT NULL,
    created_by_user_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.project_logs OWNER TO pixelcraft_user;

--
-- Name: subscriptions; Type: TABLE; Schema: public; Owner: pixelcraft_user
--

CREATE TABLE public.subscriptions (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    plan_name character varying(100) NOT NULL,
    price_per_month numeric(10,2) NOT NULL,
    status public.subscription_status DEFAULT 'ACTIVE'::public.subscription_status NOT NULL,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    next_billing_date date NOT NULL,
    canceled_at timestamp with time zone,
    plan_metadata jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    plan_id uuid,
    project_stage character varying(50) DEFAULT 'Planejamento'::character varying,
    agreed_price numeric(10,2),
    CONSTRAINT subscriptions_price_per_month_check CHECK ((price_per_month >= (0)::numeric))
);


ALTER TABLE public.subscriptions OWNER TO pixelcraft_user;

--
-- Name: TABLE subscriptions; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON TABLE public.subscriptions IS 'Assinaturas mensais de planos de desenvolvimento';


--
-- Name: COLUMN subscriptions.next_billing_date; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.subscriptions.next_billing_date IS 'Data da prÃ³xima cobranÃ§a automÃ¡tica';


--
-- Name: COLUMN subscriptions.plan_metadata; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.subscriptions.plan_metadata IS 'Features e configuraÃ§Ãµes do plano em JSON';


--
-- Name: test; Type: TABLE; Schema: public; Owner: pixelcraft_user
--

CREATE TABLE public.test (
    id integer NOT NULL,
    name character varying(50)
);


ALTER TABLE public.test OWNER TO pixelcraft_user;

--
-- Name: test_id_seq; Type: SEQUENCE; Schema: public; Owner: pixelcraft_user
--

CREATE SEQUENCE public.test_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.test_id_seq OWNER TO pixelcraft_user;

--
-- Name: test_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: pixelcraft_user
--

ALTER SEQUENCE public.test_id_seq OWNED BY public.test.id;


--
-- Name: transactions; Type: TABLE; Schema: public; Owner: pixelcraft_user
--

CREATE TABLE public.transactions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    amount numeric(15,2) NOT NULL,
    status character varying(50) DEFAULT 'pending'::character varying NOT NULL,
    type character varying(50) DEFAULT 'deposit'::character varying NOT NULL,
    provider_payment_id character varying(255),
    qr_code text,
    qr_code_base64 text,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.transactions OWNER TO pixelcraft_user;

--
-- Name: user_purchases; Type: TABLE; Schema: public; Owner: pixelcraft_user
--

CREATE TABLE public.user_purchases (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    product_id uuid NOT NULL,
    purchase_price numeric(10,2) NOT NULL,
    purchased_at timestamp with time zone DEFAULT now() NOT NULL,
    payment_id uuid,
    CONSTRAINT user_purchases_purchase_price_check CHECK ((purchase_price >= (0)::numeric))
);


ALTER TABLE public.user_purchases OWNER TO pixelcraft_user;

--
-- Name: TABLE user_purchases; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON TABLE public.user_purchases IS 'Registro de produtos comprados por cada usuÃ¡rio';


--
-- Name: COLUMN user_purchases.purchase_price; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.user_purchases.purchase_price IS 'PreÃ§o pago no momento da compra (histÃ³rico)';


--
-- Name: user_subscriptions; Type: TABLE; Schema: public; Owner: pixelcraft_user
--

CREATE TABLE public.user_subscriptions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id uuid NOT NULL,
    subscription_id uuid NOT NULL,
    payment_id uuid,
    purchased_at timestamp without time zone DEFAULT now() NOT NULL,
    expires_at timestamp without time zone NOT NULL,
    is_active boolean DEFAULT true,
    created_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.user_subscriptions OWNER TO pixelcraft_user;

--
-- Name: users; Type: TABLE; Schema: public; Owner: pixelcraft_user
--

CREATE TABLE public.users (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    username character varying(50) NOT NULL,
    email character varying(255) NOT NULL,
    password_hash character varying(255) NOT NULL,
    full_name character varying(255),
    discord_handle character varying(100),
    whatsapp_phone character varying(50),
    cpf_encrypted bytea,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    balance numeric(10,2) DEFAULT 0.00 NOT NULL,
    referral_code character varying(8),
    referred_by_code character varying(8),
    is_admin boolean DEFAULT false,
    CONSTRAINT users_balance_check CHECK ((balance >= (0)::numeric))
);


ALTER TABLE public.users OWNER TO pixelcraft_user;

--
-- Name: TABLE users; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON TABLE public.users IS 'Tabela principal de usuÃ¡rios com autenticaÃ§Ã£o e dados de perfil';


--
-- Name: COLUMN users.id; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.users.id IS 'Identificador Ãºnico UUID do usuÃ¡rio';


--
-- Name: COLUMN users.username; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.users.username IS 'Nome de usuÃ¡rio Ãºnico (login)';


--
-- Name: COLUMN users.email; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.users.email IS 'Email Ãºnico do usuÃ¡rio';


--
-- Name: COLUMN users.password_hash; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.users.password_hash IS 'Hash bcrypt da senha';


--
-- Name: COLUMN users.full_name; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.users.full_name IS 'Nome completo do usuÃ¡rio (opcional)';


--
-- Name: COLUMN users.discord_handle; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.users.discord_handle IS 'Handle do Discord (opcional)';


--
-- Name: COLUMN users.whatsapp_phone; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.users.whatsapp_phone IS 'NÃºmero do WhatsApp (opcional)';


--
-- Name: COLUMN users.cpf_encrypted; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.users.cpf_encrypted IS 'CPF criptografado com pgp_sym_encrypt (para faturamento)';


--
-- Name: COLUMN users.created_at; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.users.created_at IS 'Timestamp de criaÃ§Ã£o da conta';


--
-- Name: COLUMN users.updated_at; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.users.updated_at IS 'Timestamp da Ãºltima atualizaÃ§Ã£o';


--
-- Name: COLUMN users.balance; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.users.balance IS 'Saldo disponÃ­vel do usuÃ¡rio para compras';


--
-- Name: COLUMN users.referral_code; Type: COMMENT; Schema: public; Owner: pixelcraft_user
--

COMMENT ON COLUMN public.users.referral_code IS 'CÃ³digo de referÃªncia Ãºnico do usuÃ¡rio para indicar amigos';


--
-- Name: admin_analytics_snapshot id; Type: DEFAULT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.admin_analytics_snapshot ALTER COLUMN id SET DEFAULT nextval('public.admin_analytics_snapshot_id_seq'::regclass);


--
-- Name: test id; Type: DEFAULT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.test ALTER COLUMN id SET DEFAULT nextval('public.test_id_seq'::regclass);


--
-- Data for Name: admin_analytics_snapshot; Type: TABLE DATA; Schema: public; Owner: pixelcraft_user
--

COPY public.admin_analytics_snapshot (id, total_revenue, total_users, active_products, total_sales, last_updated, revenue_growth_pct, users_growth_pct, sales_growth_pct, products_status) FROM stdin;
1	0.00	0	0	0	2026-01-03 19:22:30.165694	0.00	-75.00	-100.00	\N
\.


--
-- Data for Name: discounts; Type: TABLE DATA; Schema: public; Owner: pixelcraft_user
--

COPY public.discounts (id, code, type, value, is_referral, created_by_user_id, expires_at, max_uses, current_uses, is_active, created_at) FROM stdin;
\.


--
-- Data for Name: files; Type: TABLE DATA; Schema: public; Owner: pixelcraft_user
--

COPY public.files (id, name, file_name, file_type, file_path, size, created_by, created_at, updated_at, is_deleted) FROM stdin;
c540e390-a597-4ab6-a399-03ab7712a329	keygen	5feb0cbc-c31d-4699-9742-9793350b03b8.jar	JAR	uploads/5feb0cbc-c31d-4699-9742-9793350b03b8.jar	199	42be7e76-d2c3-4dcf-b37c-b865946d3a2a	2025-12-14 22:40:51.572143+00	2025-12-14 22:40:51.572143+00	f
eee37d1f-3b95-4482-9f66-66a750db5cf6	image_20251125_181644c715bf91-9b0a-4e74-ad26-2407851b0650	a3346317-91ca-4799-8ca8-7122bf2dc79d.jpg	JPG	uploads/a3346317-91ca-4799-8ca8-7122bf2dc79d.jpg	312816	42be7e76-d2c3-4dcf-b37c-b865946d3a2a	2025-12-14 23:05:10.371886+00	2025-12-14 23:05:10.371886+00	f
b5d39819-16b0-4c17-88a2-62d4be673624	image_20251125_181644c715bf91-9b0a-4e74-ad26-2407851b0650	972edf09-943f-47ab-a9b7-5d06ac6da0a5.jpg	JPG	uploads/972edf09-943f-47ab-a9b7-5d06ac6da0a5.jpg	312816	42be7e76-d2c3-4dcf-b37c-b865946d3a2a	2025-12-14 23:18:57.02623+00	2025-12-14 23:18:57.02623+00	f
ed074ed5-e31e-4917-925a-1c0ef2b48008	server.jar	e252ddad-d69b-42cd-98fd-40c62ffb2c3b.zip	ZIP	uploads/e252ddad-d69b-42cd-98fd-40c62ffb2c3b.zip	1408470	42be7e76-d2c3-4dcf-b37c-b865946d3a2a	2025-12-14 23:20:55.436736+00	2025-12-14 23:20:55.436736+00	f
1c3151fd-37c8-43b0-9463-a02d057d36d4	server.jar	4a963b23-0b10-4f1e-b6ea-0f653a7b178b.zip	ZIP	uploads/4a963b23-0b10-4f1e-b6ea-0f653a7b178b.zip	1408470	42be7e76-d2c3-4dcf-b37c-b865946d3a2a	2025-12-14 23:21:51.018698+00	2025-12-14 23:21:51.018698+00	f
\.


--
-- Data for Name: library; Type: TABLE DATA; Schema: public; Owner: pixelcraft_user
--

COPY public.library (id, user_id, product_id, payment_id, purchased_at) FROM stdin;
8b959393-1c78-481c-9ef6-829c7dc01679	42be7e76-d2c3-4dcf-b37c-b865946d3a2a	7ad9e20f-7693-42e2-9aaa-3da5f116e096	da0e82f6-5ba3-42bb-807a-45693571bc84	2025-12-14 22:41:06.508746+00
\.


--
-- Data for Name: payments; Type: TABLE DATA; Schema: public; Owner: pixelcraft_user
--

COPY public.payments (id, user_id, subscription_id, description, amount, discount_applied, final_amount, status, payment_gateway_id, payment_method, payment_metadata, created_at, completed_at, failed_at) FROM stdin;
da0e82f6-5ba3-42bb-807a-45693571bc84	42be7e76-d2c3-4dcf-b37c-b865946d3a2a	\N	Purchase	0.00	0.00	0.00	COMPLETED	\N	BALANCE	\N	2025-12-14 22:41:06.510111+00	\N	\N
ba9573ac-8b11-40f3-a93f-b0af5fee122a	42be7e76-d2c3-4dcf-b37c-b865946d3a2a	\N	Purchase	0.00	0.00	0.00	COMPLETED	\N	BALANCE	\N	2025-12-14 22:49:43.101816+00	\N	\N
e0bb2a71-43f7-4753-8e10-e1a6363e7307	42be7e76-d2c3-4dcf-b37c-b865946d3a2a	\N	Purchase	0.00	0.00	0.00	COMPLETED	\N	BALANCE	\N	2025-12-14 23:05:32.749563+00	\N	\N
a3a69305-35d3-4bb5-bc5c-51fa20efd2dc	42be7e76-d2c3-4dcf-b37c-b865946d3a2a	\N	Purchase	0.00	0.00	0.00	COMPLETED	\N	BALANCE	\N	2025-12-14 23:18:34.238974+00	\N	\N
d80305e3-5ae6-44d5-aedf-16ac7adabd99	42be7e76-d2c3-4dcf-b37c-b865946d3a2a	\N	Purchase	0.00	0.00	0.00	COMPLETED	\N	BALANCE	\N	2025-12-14 23:19:09.606851+00	\N	\N
11cb0896-0b86-40ba-a992-8548730513a6	42be7e76-d2c3-4dcf-b37c-b865946d3a2a	\N	Purchase	0.00	0.00	0.00	COMPLETED	\N	BALANCE	\N	2025-12-14 23:32:45.731418+00	\N	\N
\.


--
-- Data for Name: plans; Type: TABLE DATA; Schema: public; Owner: pixelcraft_user
--

COPY public.plans (id, name, description, price, is_active, features, created_at, updated_at) FROM stdin;
e1e7a40a-e84f-42ca-8910-532746898da0	Plano Dev Pro	Para desenvolvedores que buscam ferramentas avançadas.	199.90	f	["Acesso ao Git", "CI/CD Pipeline", "Ambiente de Staging"]	2025-12-13 21:05:28.746111+00	2025-12-13 21:40:09.0633+00
f69bf0cb-58ea-4dc2-8f6c-1cf3f9a8ec83	Plano Dev Pro	Para desenvolvedores que buscam ferramentas avançadas.	199.90	f	["Acesso ao Git", "CI/CD Pipeline", "Ambiente de Staging"]	2025-12-13 21:28:14.680591+00	2025-12-13 21:40:10.919177+00
7f28b5cb-28d6-4553-a141-b340bf82b102	Plano Network	Solução completa para grandes redes de servidores.	499.90	f	["Suporte Prioritário", "Plugins Exclusivos", "Otimização de Performance"]	2025-12-13 21:05:28.746111+00	2025-12-13 21:40:12.580817+00
e90d13b2-0cd5-42d6-a84f-4954c3e1843c	Plano Network	Solução completa para grandes redes de servidores.	499.90	f	["Suporte Prioritário", "Plugins Exclusivos", "Otimização de Performance"]	2025-12-13 21:28:14.680591+00	2025-12-13 21:40:14.465899+00
fd31a11c-11a1-4c12-89d0-d911dc21759f	Plano Dev Pro	Para desenvolvedores que buscam ferramentas avançadas.	199.90	f	["Acesso ao Git", "CI/CD Pipeline", "Ambiente de Staging"]	2025-12-14 22:40:40.320068+00	2025-12-15 00:18:23.394947+00
74903014-9f00-411c-be5f-3cd5447478d3	Plano Dev Pro	Para desenvolvedores que buscam ferramentas avançadas.	199.90	f	["Acesso ao Git", "CI/CD Pipeline", "Ambiente de Staging"]	2025-12-14 22:48:33.756498+00	2025-12-15 00:18:25.282704+00
be7cfdba-bf8e-4a28-b2f0-947dbc96d5a4	Plano Dev Pro	Para desenvolvedores que buscam ferramentas avançadas.	199.90	f	["Acesso ao Git", "CI/CD Pipeline", "Ambiente de Staging"]	2025-12-14 23:19:55.727144+00	2025-12-15 00:18:27.713203+00
20ba367b-7d17-4fd5-ae6d-78dc0f39e620	Plano Network	Solução completa para grandes redes de servidores.	499.90	f	["Suporte Prioritário", "Plugins Exclusivos", "Otimização de Performance"]	2025-12-14 22:40:40.320068+00	2025-12-15 00:18:29.324189+00
bf59bfa9-9bb0-4fef-990e-8a9fc89caad3	Plano Network	Solução completa para grandes redes de servidores.	499.90	f	["Suporte Prioritário", "Plugins Exclusivos", "Otimização de Performance"]	2025-12-14 22:48:33.756498+00	2025-12-15 00:18:31.428822+00
9cf0668f-714d-4410-a92c-3fc32d99fd6c	Plano Network	Solução completa para grandes redes de servidores.	499.90	f	["Suporte Prioritário", "Plugins Exclusivos", "Otimização de Performance"]	2025-12-14 23:19:55.727144+00	2025-12-15 00:18:38.505193+00
dab6a373-72b5-4fb4-964b-1d7f0f5bfa75	Plano Básico Minecraft	Um pacote simples e eficiente, perfeito para quem está começando no mundo Minecraft. Oferecemos suporte completo para a instalação e configuração dos principais plugins, garantindo uma experiência estável.\n	139.89	f	["- Instalação do Servidor.", "- Configuração plugins básicos.(MCMMO e plugins mais complexo não está incluso)", "- Otimização Inicial.", "- Tradução de todos os plugins.", "- Suporte via ticket no discord."]	2025-12-15 04:40:25.400563+00	2025-12-15 04:42:07.582805+00
7a08e6ba-8ba8-4a2f-a8c0-6a9f748efb43	Plano básico Minecraft	Um pacote simples e eficiente, perfeito para quem está começando no mundo Minecraft. \n\nOferecemos suporte completo para a instalação e configuração dos principais plugins, garantindo uma experiência estável.	139.90	f	["Instalação do Servidor.", "Configuração plugins básicos.", "Otimização Inicial.", "Tradução de todos os plugins.", "Suporte via ticket no discord."]	2025-12-15 04:46:38.335719+00	2025-12-15 04:47:20.926774+00
f35bb574-829a-4b89-917c-146868cf367c	Plano Avançado Minecraft	Configuração completa do servidor com plugins confiáveis e reconhecidos no mercado, focada em desempenho, estabilidade e segurança. Inclui otimização avançada para redução de lag, correção de bugs, resolução de conflitos e suporte técnico prioritário 24h via Discord.	299.90	f	["- Configuração e tradução de plugins avançados.", "- Configuração de Mods.", "- Instalação de plugins de desempenho e jogabilidade.", "- Otimização avançada para reduzir lag.", "- Suporte técnico prioritário 24h via Discord.", "- Correção de ResourcePack / ItemsAdder / Oraxen.", "- Correção de Bugs e conflitos no servidor.", "- Análise do servidor para identificar melhorias."]	2025-12-15 05:04:45.580589+00	2025-12-15 05:05:48.258678+00
6a044d28-a5f2-493e-9139-ba3d046a87fb	Plano básico Minecraft	Um pacote simples e eficiente, perfeito para quem está começando no mundo Minecraft. Oferecemos suporte completo para a instalação e configuração dos principais plugins, garantindo uma experiência estável.	139.90	f	["- Instalação do Servidor.", "- Configuração plugins básicos.", "- Otimização Inicial.", "- Tradução de todos os plugins.", "- Suporte via ticket no discord."]	2025-12-15 04:53:39.162053+00	2025-12-15 05:08:45.344898+00
ce4f2dfc-d1ec-44c7-981a-31aeccdfb380	Plano Network Minecraft	O Plano Network é ideal para quem deseja construir e expandir uma rede de servidores, atendendo múltiplas modalidades com estrutura sólida, escalável e preparada para crescer no mercado.	359.90	f	["- Configuração de Proxy.", "- Integração de servidores.", "- Configuração de Plugins.", "- Suporte prioritário e 24/7 via ticket no discord.", "- Instalação de Servidores.", "- Tradução de Servidores.", "- Otimização completa."]	2025-12-15 05:16:57.632691+00	2025-12-15 05:17:46.564023+00
b4d5a6ba-33fe-4f67-a6c8-0339075e1c01	Plano Premium Ragnarok	O Plano Premium Ragnarok é voltado para quem busca um servidor profissional, otimizado e pronto para competir no mercado. Oferecemos um serviço completo, com desenvolvimento avançado, identidade visual, estabilidade, segurança e suporte técnico prioritário.	349.90	t	["Tudo do Plano Básico.", "Criação e configuração de patcher automático.", "Desenvolvimento e edição de scripts personalizados.", "Criação de missões exclusivas e eventos automatizados.", "Configuração avançada de balanceamento (PvP, WoE e economia).", "Tradução e adaptação completa do cliente.", "Configuração de site do servidor (base ou integração).", "Otimização de desempenho e redução de lag.", "Análise de segurança e prevenção contra exploits.", "Suporte técnico prioritário 24/7 via Discord/Whatsapp."]	2026-01-03 18:25:33.349227+00	2026-01-03 18:41:59.480806+00
0108668f-2888-481f-a8b0-4fab978eb19d	Plano Avançado Minecraft	Ideal para servidores que buscam alto desempenho, estabilidade e profissionalismo.\n\nEste plano oferece uma configuração completa e otimizada, utilizando plugins confiáveis e amplamente conhecidos no mercado, garantindo segurança, compatibilidade e uma experiência de jogo estável.\n\nInclui instalação e configuração de plugins de desempenho e jogabilidade, otimização avançada para redução de lag, correção de bugs e resolução de conflitos entre sistemas, além de uma análise técnica detalhada do servidor para identificar melhorias e ajustes necessários. O plano também conta com suporte técnico prioritário 24h via Discord, assegurando rapidez no atendimento e acompanhamento contínuo do servidor.\n\n🔹 Recomendado para servidores em crescimento ou já consolidados que desejam elevar o padrão de qualidade e confiabilidade.	299.90	f	["- Configuração e tradução de plugins avançados.", "- Instalação de plugins de desempenho e jogabilidade.", "- Otimização avançada para reduzir lag.", "- Suporte técnico prioritário 24h via Discord.", "- Correção de ResourcePack / ItemsAdder / Oraxen.", "- Correção de Bugs e conflitos no servidor.", "- Análise do servidor para identificar melhorias."]	2025-12-15 05:02:50.165286+00	2025-12-15 05:04:52.034516+00
1b56aed9-34a4-4013-8948-142bba471a17	Plano Básico Minecraft	Um pacote simples e eficiente, perfeito para quem está começando no mundo Minecraft. Oferecemos suporte completo para a instalação e configuração dos principais plugins, garantindo uma experiência estável.	139.90	t	["- Instalação do Servidor.", "- Instalação de Mods.", "- Configuração plugins básicos.", "- Otimização Inicial.", "- Tradução de todos os plugins.", "- Suporte via ticket no discord."]	2025-12-15 05:08:39.79856+00	2025-12-15 05:08:39.79856+00
e374dc8c-d07b-4270-969b-b05dc8c1b4dd	Plano Avançado Minecraft	Configuração completa do servidor com plugins confiáveis e reconhecidos no mercado, focada em desempenho, estabilidade e segurança. Inclui otimização avançada para redução de lag, correção de bugs, resolução de conflitos e suporte técnico prioritário 24h via Discord.	239.90	f	["- Configuração e tradução de plugins avançados.", "- Configuração de Mods. (Ex: DBC/Pixelmon etc...)", "- Instalação de plugins de desempenho e jogabilidade.", "- Otimização avançada para reduzir lag.", "- Suporte técnico prioritário 24h via Discord.", "- Correção de ResourcePack / ItemsAdder / Oraxen.", "- Correção de Bugs e conflitos no servidor.", "- Análise do servidor para identificar melhorias."]	2025-12-15 05:06:47.028541+00	2025-12-28 19:30:00.614507+00
5ffa8f43-af35-467e-b624-62a265c5a15a	Plano Avançado Minecraft	O plano completo para quem busca o máximo em personalização e desempenho. Inclui configurações detalhadas, otimizações extremas e suporte contínuo.	239.90	t	["Configuração completa e personalizada de plugins e mods.", "Otimização máxima para eliminar lag.", "Suporte técnico prioritário 24h/dia via Discord.", "Introdução de ResourcePack.", "Suporte em texturas.", "Análise contínua de desempenho e segurança para garantir a máxima proteção e eficiência."]	2025-12-28 19:32:33.008348+00	2025-12-28 19:32:33.008348+00
e2c02d67-b270-45e0-8ffe-5ca923eb606c	Plano Network Minecraft	O Plano Network é ideal para quem deseja construir e expandir uma rede de servidores, atendendo múltiplas modalidades com estrutura sólida, escalável e preparada para crescer no mercado.	359.90	f	["- Configuração de Proxy.", "- Integração de servidores.", "- Configuração de Plugins.", "- Suporte prioritário e 24/7 via ticket no discord.", "- Instalação de Servidores.", "- Tradução de Servidores.", "- Correção de bugs/problemas nos servidores.", "- Otimização completa."]	2025-12-15 05:18:51.03465+00	2025-12-28 19:33:01.367341+00
60997d2c-468c-42ea-b7fc-f320e56be222	Plano Network Minecraft	O Plano Network é ideal para quem deseja construir e expandir uma rede de servidores, atendendo múltiplas modalidades com estrutura sólida, escalável e preparada para crescer no mercado.\n\n	299.90	t	["Configuração de Proxy.", "Integração de servidores.", "Configuração de Plugins.", "Suporte prioritário e 24/7 via ticket no discord.", "Instalação de Servidores.", "Tradução de Servidores.", "Correção de bugs/problemas nos servidores.", "Otimização completa."]	2025-12-28 19:33:55.255604+00	2025-12-28 19:33:55.255604+00
9c8ba0c3-0c6a-4a95-82c6-c7729a83f3d6	Plano Sócio Minecraft	Configuração completa dos servidores com plugins confiáveis e reconhecidos no mercado, focada em desempenho, estabilidade e segurança. Inclui otimização avançada para redução de lag, correção de bugs, resolução de conflitos e suporte técnico prioritário 24h via Discord.	479.90	t	["Serviço completo e sob medida", "Configuração de plugins e mods", "Testes de desempenho contínuos", "Análise de segurança", "Consultoria estratégica para seu servidor", "e tudo relacionado aos planos anteriores."]	2025-12-28 19:35:56.046247+00	2025-12-28 19:35:56.046247+00
f2f312b9-7aaf-41aa-b764-7a1bbd3d6d78	Plano Especial	asdasdsadas	119.90	f	["12321321asdasdas", "21", "3", "123", "21", "e", "21"]	2025-12-28 19:53:27.709919+00	2025-12-28 19:53:37.167656+00
1ff38fb2-5c25-461e-9e63-9dd9739d7a4c	tsrwet	asdf	120.00	f	["sedfasd", "asf", "sda"]	2025-12-28 20:15:32.597979+00	2025-12-28 20:15:38.012078+00
5e2a1a1a-31fa-4ba6-b39b-4a1e4a7c390e	Plano Básico Ragnarok	O Plano Básico Ragnarok é ideal para quem deseja iniciar ou organizar um servidor com qualidade, estabilidade e identidade própria. Contamos com profissionais especializados em Ragnarok Online para configurar e estruturar os principais sistemas do servidor, garantindo uma base sólida para crescimento.	149.90	t	["Instalação e configuração do emulador (rAthena / Hercules).", "Configuração inicial do servidor (rates, drops, classes e mapas).", "Tradução completa do servidor (itens, NPCs, skills e mensagens).", "Configuração de NPCs essenciais (warper, healer, reset, jobs).", "Correções básicas de scripts e ajustes de funcionamento.", "Configuração do cliente (GRF básica).", "Suporte técnico via ticket no Discord."]	2026-01-03 18:25:12.283014+00	2026-01-03 18:41:54.069549+00
3424c7f2-418d-46cb-a0f6-9e9b3db92cd9	dsfasdf	dfasd	1000.00	f	["sdfasfdas", "asdf", "asfds"]	2025-12-28 20:22:52.118002+00	2025-12-28 20:23:07.128802+00
543193b1-6bff-476f-bbc5-e618aeb32789	Plano Dev Pro 2	Para desenvolvedores que buscam ferramentas avançadas.	199.90	f	["Acesso ao Git", "CI/CD Pipeline", "Ambiente de Staging"]	2025-12-28 20:19:09.136738+00	2026-01-03 18:24:29.426097+00
8eaebdcc-09fe-44cd-a937-5478e88dfcfc	Plano Network	Solução completa para grandes redes de servidores.	499.90	f	["Suporte Prioritário", "Plugins Exclusivos", "Otimização de Performance"]	2025-12-28 20:19:09.136738+00	2026-01-03 18:24:34.747061+00
\.


--
-- Data for Name: products; Type: TABLE DATA; Schema: public; Owner: pixelcraft_user
--

COPY public.products (id, name, description, price, type, download_url_encrypted, is_exclusive, stock_quantity, image_url, is_active, created_at, updated_at) FROM stdin;
0e9e8710-3a98-428b-9041-62db3677b831	RPG Vol 1 - Elemental Altars	✨ RPG Vol. 1 – Altares Elementais | Addon para ItemsAdder ✨​\n\nRPG Vol. 1 – Altares Elementais é um addon de alta qualidade para ItemsAdder que apresenta 8 altares místicos, cada um representando os poderes elementais da Luz e das Trevas nos quatro elementos principais: Fogo, Água, Terra e Ar.\n\n🌟 O que está incluído:\n4 Altares Elementais da Luz: Fogo, Água, Terra, Ar\n4 Altares Elementais das Trevas: Fogo, Água, Terra, Ar\nCada altar é um item/bloco 3D totalmente personalizado e texturizado, pronto para adicionar profundidade e magia ao seu servidor de RPG.\n\n📦 Formato do pacote:\nAddon ItemsAdder completo\nOrganizado e pronto para instalar\nLeve e com ótimo desempenho\n\n🎮 Detalhes técnicos:\n100% compatível com ItemsAdder\nInclui todos os arquivos necessários: texturas, modelos, itens, modelos de blocos, imagens de fontes e idiomas\n\nInstalação fácil por arrastar e soltar\nNão substitui nenhuma textura original do jogo.\nNão requer OptiFine!	39.90	TEXTUREPACK	\\x	f	9999	https://builtbybit.com/attachments/12312312312321-png.877340/?preset=fullr1	f	2025-12-15 00:22:32.709816+00	2025-12-15 00:22:41.821184+00
5fdd8f83-8a3f-4c4f-9b81-96ada5c133d1	RPG Vol 1 - Elemental Altars	✨ RPG Vol. 1 – Altares Elementais | Addon para ItemsAdder ✨​\n\nRPG Vol. 1 – Altares Elementais é um addon de alta qualidade para ItemsAdder que apresenta 8 altares místicos, cada um representando os poderes elementais da Luz e das Trevas nos quatro elementos principais: Fogo, Água, Terra e Ar.\n\n🌟 O que está incluído:\n4 Altares Elementais da Luz: Fogo, Água, Terra, Ar\n4 Altares Elementais das Trevas: Fogo, Água, Terra, Ar\nCada altar é um item/bloco 3D totalmente personalizado e texturizado, pronto para adicionar profundidade e magia ao seu servidor de RPG.\n\n📦 Formato do pacote:\nAddon ItemsAdder completo\nOrganizado e pronto para instalar\nLeve e com ótimo desempenho\n\n🎮 Detalhes técnicos:\n100% compatível com ItemsAdder\nInclui todos os arquivos necessários: texturas, modelos, itens, modelos de blocos, imagens de fontes e idiomas\n\nInstalação fácil por arrastar e soltar\nNão substitui nenhuma textura original do jogo.\nNão requer OptiFine!	39.90	TEXTUREPACK	\\x	f	9999	https://builtbybit.com/attachments/12312312312321-png.877340/?preset=fullr1	f	2025-12-15 00:22:32.702709+00	2025-12-15 00:22:45.388267+00
6425a5b7-9b37-41ef-8a8c-73f42862c4b0	Age Medieval Pack Vol. 1	Age Medieval Pack é um pacote de texturas de armaduras com tema medieval, contendo mais de 200 modelos exclusivos.	69.90	TEXTUREPACK	\\x	t	9999	https://builtbybit.com/attachments/medieval_pack-png.766702/?preset=fullr1	t	2025-12-14 22:01:15.683008+00	2026-01-03 19:18:36.091392+00
f28c17ae-9bdc-4a25-8f71-a5f34029256b	RPG Vol 1 - Elemental Altars	✨ RPG Vol. 1 – Altares Elementais | Addon para ItemsAdder ✨​\n\nRPG Vol. 1 – Altares Elementais é um addon de alta qualidade para ItemsAdder que apresenta 8 altares místicos, cada um representando os poderes elementais da Luz e das Trevas nos quatro elementos principais: Fogo, Água, Terra e Ar.\n\n🌟 O que está incluído:\n4 Altares Elementais da Luz: Fogo, Água, Terra, Ar\n4 Altares Elementais das Trevas: Fogo, Água, Terra, Ar\nCada altar é um item/bloco 3D totalmente personalizado e texturizado, pronto para adicionar profundidade e magia ao seu servidor de RPG.\n\n📦 Formato do pacote:\nAddon ItemsAdder completo\nOrganizado e pronto para instalar\nLeve e com ótimo desempenho\n\n🎮 Detalhes técnicos:\n100% compatível com ItemsAdder\nInclui todos os arquivos necessários: texturas, modelos, itens, modelos de blocos, imagens de fontes e idiomas\n\nInstalação fácil por arrastar e soltar\nNão substitui nenhuma textura original do jogo.\nNão requer OptiFine!	39.90	TEXTUREPACK	\\x	t	9999	https://builtbybit.com/attachments/12312312312321-png.877340/?preset=fullr1	t	2025-12-15 00:22:16.697671+00	2026-01-03 19:18:40.210863+00
7ad9e20f-7693-42e2-9aaa-3da5f116e096	Teste	Fdroid	0.00	PLUGIN	\\x	f	94		f	2025-12-14 20:07:12.609565+00	2025-12-14 23:21:00.910043+00
1b52488c-a556-4e02-affd-00ec3a131125	sdg	wdgs	0.00	PLUGIN	\\x	f	998		f	2025-12-14 23:32:31.235659+00	2025-12-15 00:14:07.645482+00
2c3b2476-81da-414f-bc02-1821af9879b5	Servidor Survival (1.21+)	 	399.90	SERVER_TEMPLATE	\\x	t	10		f	2025-12-14 21:54:55.437939+00	2025-12-15 00:14:13.749721+00
cde5ec57-8409-4207-a833-183d6546fe33	RPG Vol 1 - Elemental Altars	✨ RPG Vol. 1 – Altares Elementais | Addon para ItemsAdder ✨​\n\nRPG Vol. 1 – Altares Elementais é um addon de alta qualidade para ItemsAdder que apresenta 8 altares místicos, cada um representando os poderes elementais da Luz e das Trevas nos quatro elementos principais: Fogo, Água, Terra e Ar.\n\n🌟 O que está incluído:\n4 Altares Elementais da Luz: Fogo, Água, Terra, Ar\n4 Altares Elementais das Trevas: Fogo, Água, Terra, Ar\nCada altar é um item/bloco 3D totalmente personalizado e texturizado, pronto para adicionar profundidade e magia ao seu servidor de RPG.\n\n📦 Formato do pacote:\nAddon ItemsAdder completo\nOrganizado e pronto para instalar\nLeve e com ótimo desempenho\n\n🎮 Detalhes técnicos:\n100% compatível com ItemsAdder\nInclui todos os arquivos necessários: texturas, modelos, itens, modelos de blocos, imagens de fontes e idiomas\n\nInstalação fácil por arrastar e soltar\nNão substitui nenhuma textura original do jogo.\nNão requer OptiFine!	39.90	TEXTUREPACK	\\x	f	9999	https://builtbybit.com/attachments/12312312312321-png.877340/?preset=fullr1	f	2025-12-15 00:22:19.430076+00	2025-12-15 00:22:23.684672+00
4fd4eee1-3cea-4e06-a16a-9fd7cf9e1f14	Wild Island	🗺️ Mapa: Wild Island Lobby\n✏️ Dimensões: 250x250\n📍 Versões: Java 1.8x - 1.20x\n📌 Tema: Lobby	49.90	MAP	\\x	f	9999	https://i.imgur.com/lqpiLdW.png	t	2026-01-03 18:57:33.053761+00	2026-01-03 18:57:33.053761+00
8f36cf9e-54f6-41dd-82cc-7c59924363fb	Minecraft SkyBlock (1.21+)	 	249.90	SERVER_TEMPLATE	\\x	f	9999		t	2025-12-15 05:32:23.811707+00	2026-01-03 19:15:21.893803+00
ca17358f-eaf0-47f4-b6c9-a0ef6b8e3603	Minecraft Survival Vanilla (1.21+)	 	249.90	SERVER_TEMPLATE	\\x	f	9999		t	2025-12-15 00:35:58.040906+00	2026-01-03 19:11:07.36942+00
5a4ccadc-0bc1-42e6-86a6-0715adaa2221	Servidor de FiveM Americano	asdsadasdasda	3000.00	PLUGIN	\\x	t	1		f	2025-12-21 19:53:08.499506+00	2025-12-28 19:28:45.551274+00
e2a874b9-2319-443e-8ddd-b447f52d426e	Thunderbolt Animated Weapons	✏️ Nome: Thunderbolt Animated Weapons\n\n📍 Versões: 1.14+\n\n📦Incluído no pacote:\n\nWeapons:\n\n– 1x Sword\n– 1x Scythe\n– 1x Bows (3 Steps)\n– 1x Shields (2 Steps)\n– 1x Spear\n– 1x Hammer\n– 1x Staff\n– 1x Crossbow\n– 1x Greatsword\n\n(Os materiais para os itens de espada incluem Lança, Martelo, Foice, Cajado e Tridente.)\n\nTools:\n\n– 1x Axe\n– 1x Hoe\n– 1x Pickaxe\n– 1x Shovel\n– 1x Fishing Rod (2 Steps)\n\nEquipment:\n\n– Armor (32x Helmet, Chestplate, Leggings and Boots)\n– 3D Helmet\n– Wings\n\nOther:\n\n– Chest\n– Key	67.79	TEXTUREPACK	\\x	f	9999	https://i.imgur.com/Axh5c7I.png	t	2026-01-03 19:07:16.774108+00	2026-01-03 19:07:56.334663+00
2639ace6-ad68-42f0-ba49-bf64210119ab	Tropical Isle	🗺️ Mapa: Tropical Isle\n✏️ Dimensões: 350x350\n📍 Versões: Java 1.13+\n📌 Tema: Lobby	9.90	MAP	\\x	f	9999	https://i.imgur.com/L54tQAj.png	t	2026-01-03 18:51:37.74712+00	2026-01-03 18:51:37.74712+00
291c839e-7b0f-44da-a4b9-dc1399f53b99	Verdant Valley	🗺️ Mapa: Verdant Valley\n✏️ Dimensões: 4000 x 4000\n📍 Versões: Java 1.16.5, Java 1.20.1 e Bedrock 1.20\n📌 Tema: Lobby	24.90	MAP	\\x	f	9999	https://i.imgur.com/t5bfhHx.png	t	2026-01-03 18:45:23.013753+00	2026-01-03 18:51:55.573077+00
91019a4e-d855-42d3-9bb6-04406293f5fd	Ancient Greece	🗺️ Mapa: Ancient Greece\n✏️ Dimensões: 245x245\n📍 Versões: Java 1.13+\n📌 Tema: Lobby	49.90	MAP	\\x	f	9999	https://i.imgur.com/pVufvUW.png	t	2026-01-03 18:53:27.682356+00	2026-01-03 18:53:27.682356+00
165948fb-3a6c-40d0-8111-c251ae932f37	Orient Lobby	🗺️ Mapa: Orient Lobby\n✏️ Dimensões: 500x500\n📍 Versões: Java 1.8x - 1.20x\n📌 Tema: Lobby	14.90	MAP	\\x	f	9999	https://i.imgur.com/aWgFfmc.png	t	2026-01-03 18:55:07.657077+00	2026-01-03 18:55:07.657077+00
ba7e3276-53d2-468e-a943-e5c4bbad7acf	Christmas Weapon Animated Set	✏️ Nome: Christmas Weapon Animated Set\n\n📍 Versões: 1.14+ \n\n📦Incluído no pacote Christmas Weapon Animated Set:\n\n- 1x Axe\n\n- 1x Bigsword\n\n- 1x Bow\n\n- 1x Cane\n\n- 1x Club\n\n- 1x Christmastree\n\n- 1x Dagger\n\n- 1x Fishing Rod\n\n- 1x Halberd\n\n- 1x Hammer\n\n- 1x Hat\n\n- 1x Hoe\n\n- 1x Key\n\n- 1x Pickaxe\n\n- 1x Shield\n\n- 1x Shovel\n\n- 1x Spear\n\n- 1x Staff\n\n- 1x Sword\n\n- 1x Wings\n\n- 1x Chest\n\n- 1x Helmet\n\n- 1x Chestplate\n\n- 1x Leggings\n\n- 1x Boots\n\nConfigurações pré-definidas.\n\nConfiguração de arrastar e soltar do ItemsAdder.\n\nConfiguração de arrastar e soltar do Oraxen.\n\nConfiguração de arrastar e soltar do Nexo.\n\nConfiguração de arrastar e soltar do CraftEngine.\n\nPacote de recursos do jogo base.	67.79	TEXTUREPACK	\\x	f	9999	https://i.imgur.com/Vh48Sfn.png	t	2026-01-03 19:04:12.743017+00	2026-01-03 19:09:17.56444+00
8275edbb-7959-4703-8277-15e5273cab2e	Minecraft Survival Custom [1.21+]	 	599.90	SERVER_TEMPLATE	\\x	f	9999		t	2025-12-15 00:38:12.626059+00	2026-01-03 19:11:43.73621+00
f5a42025-e655-49ee-b4fc-d78a4ce402dd	Minecraft Bedwars (1.8+)	 	249.90	SERVER_TEMPLATE	\\x	f	9999		t	2025-12-15 05:28:03.706539+00	2026-01-03 19:12:38.825678+00
973b6732-9b99-4fc0-ba2c-f39576b5965b	Minecraft BoxPvP (1.21+)	 	299.90	SERVER_TEMPLATE	\\x	f	9999		t	2025-12-15 05:47:29.618488+00	2026-01-03 19:16:02.840006+00
c63ca84f-1160-4ec5-9ebc-3fa5aeabec27	Minecraft Pixelmon (1.12.2)	 	399.90	SERVER_TEMPLATE	\\x	f	9999		t	2025-12-15 05:28:25.072994+00	2026-01-03 19:12:59.307255+00
9a1fbf4c-efc8-4e28-9c62-7577b91ae3b8	Minecraft Pixelmon (1.16.5)	 	599.90	SERVER_TEMPLATE	\\x	f	9999		t	2025-12-15 05:28:44.086175+00	2026-01-03 19:13:19.085665+00
8c322d58-d8d9-4398-9393-0cb484453644	Minecraft DragonBlock (1.7.10)	 	399.90	SERVER_TEMPLATE	\\x	f	9999		t	2025-12-15 05:29:28.528507+00	2026-01-03 19:13:57.841263+00
b78298e1-0f41-4945-93f4-a004f88a967b	Minecraft KitPvP (1.8)	 	199.90	SERVER_TEMPLATE	\\x	f	9999		t	2025-12-15 05:30:43.793632+00	2026-01-03 19:14:26.066488+00
9aa26bcc-be3e-45ab-8445-4290d7582c62	Minecraft Factions (1.21+)	 	199.90	SERVER_TEMPLATE	\\x	f	9999		t	2025-12-15 05:33:26.800585+00	2026-01-03 19:15:11.357679+00
d5444e50-2185-49b8-961a-3ae9081d06a9	Plugin de Altar (1.21 apenas)	Descrição:\n\n\n\n- Crie infinitos altares.\n- Configure o tempo de cada altar.\n- Configure os itens de cada altar.\n- Configure a quantidade de drops por tempo do altar.\n- Menus configuráveis.\n- Menu para gerenciar todos os altares.\n\n\nComandos:\n\n\n\n\nDependências:\n\n- \n- 	29.90	PLUGIN	\\x	t	9999		t	2025-12-15 00:48:08.84337+00	2026-01-03 19:18:03.857858+00
c1039cf6-512f-4f7c-8138-a5094b5b89a6	Ores Craft Armor Pack - Vol 1	Uma coleção exclusiva de texturas de armadura inspiradas no icônico mod OreSpawn, criada para trazer personalização e estilo ao seu servidor de Minecraft!\n\nPrincipais Características:\n\nTexturas Premium: Inclui 14 designs de armadura exclusivos, cada um com detalhes vibrantes e um visual impressionante.\n\nCompatibilidade Total: Configurações otimizadas para ItemsAdder, permitindo a integração perfeita das texturas ao seu servidor.	59.90	TEXTUREPACK	\\x	t	9999	https://builtbybit.com/attachments/image_modified-png.865831/?preset=fullr1	t	2025-12-15 00:24:19.495044+00	2026-01-03 19:18:44.140491+00
\.


--
-- Data for Name: project_logs; Type: TABLE DATA; Schema: public; Owner: pixelcraft_user
--

COPY public.project_logs (id, subscription_id, message, created_by_user_id, created_at) FROM stdin;
\.


--
-- Data for Name: subscriptions; Type: TABLE DATA; Schema: public; Owner: pixelcraft_user
--

COPY public.subscriptions (id, user_id, plan_name, price_per_month, status, started_at, next_billing_date, canceled_at, plan_metadata, created_at, updated_at, plan_id, project_stage, agreed_price) FROM stdin;
\.


--
-- Data for Name: test; Type: TABLE DATA; Schema: public; Owner: pixelcraft_user
--

COPY public.test (id, name) FROM stdin;
\.


--
-- Data for Name: transactions; Type: TABLE DATA; Schema: public; Owner: pixelcraft_user
--

COPY public.transactions (id, user_id, amount, status, type, provider_payment_id, qr_code, qr_code_base64, created_at, updated_at) FROM stdin;
df11c87f-b00b-4e87-9710-ca2ccc4012ef	42be7e76-d2c3-4dcf-b37c-b865946d3a2a	1.10	pending	deposit	137144929097	\N	\N	2025-12-13 21:16:14.052131+00	2025-12-13 21:16:14.052131+00
9b3300a0-44c7-4227-982d-2482cda85143	42be7e76-d2c3-4dcf-b37c-b865946d3a2a	1.01	pending	deposit	137144691295	\N	\N	2025-12-13 21:16:55.564942+00	2025-12-13 21:16:55.564942+00
\.


--
-- Data for Name: user_purchases; Type: TABLE DATA; Schema: public; Owner: pixelcraft_user
--

COPY public.user_purchases (id, user_id, product_id, purchase_price, purchased_at, payment_id) FROM stdin;
d34601df-a128-4912-8ffc-cfbaba461fb5	42be7e76-d2c3-4dcf-b37c-b865946d3a2a	7ad9e20f-7693-42e2-9aaa-3da5f116e096	0.00	2025-12-14 22:49:43.099371+00	ba9573ac-8b11-40f3-a93f-b0af5fee122a
5d8656e4-d140-438d-bc47-02b7fb8992e3	42be7e76-d2c3-4dcf-b37c-b865946d3a2a	1b52488c-a556-4e02-affd-00ec3a131125	0.00	2025-12-14 23:32:45.724913+00	11cb0896-0b86-40ba-a992-8548730513a6
\.


--
-- Data for Name: user_subscriptions; Type: TABLE DATA; Schema: public; Owner: pixelcraft_user
--

COPY public.user_subscriptions (id, user_id, subscription_id, payment_id, purchased_at, expires_at, is_active, created_at) FROM stdin;
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: pixelcraft_user
--

COPY public.users (id, username, email, password_hash, full_name, discord_handle, whatsapp_phone, cpf_encrypted, created_at, updated_at, balance, referral_code, referred_by_code, is_admin) FROM stdin;
dfa02cd0-9880-44ee-be7f-28d59a69487f	mrmadara667	thales667@hotmail.com	$2a$10$lOYI6CmWuw2Yu/B13ETs0udskgCmQ.Md2hqt5gx7YBgpvREvI8LDK	Thales dos Santos	mrmadara667		\N	2025-12-14 04:06:56.544725+00	2025-12-14 15:22:37.700355+00	0.00	UI9THJ4K	\N	t
42be7e76-d2c3-4dcf-b37c-b865946d3a2a	pedrortxdev	dsleal298@gmail.com	$2a$10$hfBNgmSUGM7dR.qK4CASMe9E9WKwkuGgXQvgirlOnjA5HYw.qlcpS	daniel leal	pedreirodev	55984451001	\N	2025-12-13 02:03:04.493937+00	2025-12-14 23:32:45.724913+00	0.00	TPE0Q17L	\N	t
eb2927d7-5fea-495a-bab1-f9d504892eb6	natalino	davidmiranda007007@gmail.com	$2a$10$U3ABoGaAlRfKAH5fg73uWuEsFoVuOj03./g.aGgX.LPZex0h.OjGa	david gabriel miranda de souza			\N	2025-12-15 05:00:37.82512+00	2025-12-15 05:00:37.82512+00	0.00	SZNUDVDL	\N	f
a65d323c-7194-48b9-87c0-5d013d38e127	hahahuhu	hahahuhu9298@gmail.com	$2a$10$6H0QKXfsB9LZyX9TbI7QmugU1MoDErnetgHmZRjH0ITBkzqCmWWLq	peter parker			\N	2025-12-20 23:43:37.017671+00	2025-12-20 23:43:37.017671+00	0.00	601H2O9V	\N	f
964292af-b2c5-4642-a4fb-9e355371ade2	mysrael	mysraelbraga775@gmail.com	$2a$10$IUKtWvUJ/x8qVngBvsmuquYIPgG3tNrNtfZEbu9PNR5NfFn/743Va	mysrael braga coutinho			\N	2026-01-03 19:22:18.653336+00	2026-01-03 19:22:18.653336+00	0.00	5NGOUXOA	\N	f
\.


--
-- Name: admin_analytics_snapshot_id_seq; Type: SEQUENCE SET; Schema: public; Owner: pixelcraft_user
--

SELECT pg_catalog.setval('public.admin_analytics_snapshot_id_seq', 1, false);


--
-- Name: test_id_seq; Type: SEQUENCE SET; Schema: public; Owner: pixelcraft_user
--

SELECT pg_catalog.setval('public.test_id_seq', 1, false);


--
-- Name: admin_analytics_snapshot admin_analytics_snapshot_pkey; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.admin_analytics_snapshot
    ADD CONSTRAINT admin_analytics_snapshot_pkey PRIMARY KEY (id);


--
-- Name: discounts discounts_code_key; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.discounts
    ADD CONSTRAINT discounts_code_key UNIQUE (code);


--
-- Name: discounts discounts_pkey; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.discounts
    ADD CONSTRAINT discounts_pkey PRIMARY KEY (id);


--
-- Name: files files_pkey; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.files
    ADD CONSTRAINT files_pkey PRIMARY KEY (id);


--
-- Name: library library_pkey; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.library
    ADD CONSTRAINT library_pkey PRIMARY KEY (id);


--
-- Name: payments payments_pkey; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT payments_pkey PRIMARY KEY (id);


--
-- Name: plans plans_pkey; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.plans
    ADD CONSTRAINT plans_pkey PRIMARY KEY (id);


--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);


--
-- Name: project_logs project_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.project_logs
    ADD CONSTRAINT project_logs_pkey PRIMARY KEY (id);


--
-- Name: subscriptions subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_pkey PRIMARY KEY (id);


--
-- Name: test test_pkey; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.test
    ADD CONSTRAINT test_pkey PRIMARY KEY (id);


--
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);


--
-- Name: library unique_user_product; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.library
    ADD CONSTRAINT unique_user_product UNIQUE (user_id, product_id);


--
-- Name: user_purchases user_purchases_pkey; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.user_purchases
    ADD CONSTRAINT user_purchases_pkey PRIMARY KEY (id);


--
-- Name: user_purchases user_purchases_user_id_product_id_key; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.user_purchases
    ADD CONSTRAINT user_purchases_user_id_product_id_key UNIQUE (user_id, product_id);


--
-- Name: user_subscriptions user_subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.user_subscriptions
    ADD CONSTRAINT user_subscriptions_pkey PRIMARY KEY (id);


--
-- Name: user_subscriptions user_subscriptions_user_id_subscription_id_key; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.user_subscriptions
    ADD CONSTRAINT user_subscriptions_user_id_subscription_id_key UNIQUE (user_id, subscription_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_referral_code_key; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_referral_code_key UNIQUE (referral_code);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: idx_discounts_active; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_discounts_active ON public.discounts USING btree (is_active);


--
-- Name: idx_discounts_code; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_discounts_code ON public.discounts USING btree (code);


--
-- Name: idx_discounts_creator; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_discounts_creator ON public.discounts USING btree (created_by_user_id);


--
-- Name: idx_files_created_by; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_files_created_by ON public.files USING btree (created_by);


--
-- Name: idx_files_created_by_deleted; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_files_created_by_deleted ON public.files USING btree (created_by, is_deleted);


--
-- Name: idx_files_is_deleted; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_files_is_deleted ON public.files USING btree (is_deleted);


--
-- Name: idx_library_product; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_library_product ON public.library USING btree (product_id);


--
-- Name: idx_library_user; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_library_user ON public.library USING btree (user_id);


--
-- Name: idx_payments_created_at; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_payments_created_at ON public.payments USING btree (created_at DESC);


--
-- Name: idx_payments_gateway_id; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_payments_gateway_id ON public.payments USING btree (payment_gateway_id);


--
-- Name: idx_payments_status; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_payments_status ON public.payments USING btree (status);


--
-- Name: idx_payments_subscription; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_payments_subscription ON public.payments USING btree (subscription_id);


--
-- Name: idx_payments_user; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_payments_user ON public.payments USING btree (user_id);


--
-- Name: idx_products_active; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_products_active ON public.products USING btree (is_active);


--
-- Name: idx_products_price; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_products_price ON public.products USING btree (price);


--
-- Name: idx_products_type; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_products_type ON public.products USING btree (type);


--
-- Name: idx_project_logs_subscription_id; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_project_logs_subscription_id ON public.project_logs USING btree (subscription_id);


--
-- Name: idx_subscriptions_billing_date; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_subscriptions_billing_date ON public.subscriptions USING btree (next_billing_date);


--
-- Name: idx_subscriptions_plan_id; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_subscriptions_plan_id ON public.subscriptions USING btree (plan_id);


--
-- Name: idx_subscriptions_status; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_subscriptions_status ON public.subscriptions USING btree (status);


--
-- Name: idx_subscriptions_user; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_subscriptions_user ON public.subscriptions USING btree (user_id);


--
-- Name: idx_transactions_user_id; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_transactions_user_id ON public.transactions USING btree (user_id);


--
-- Name: idx_user_purchases_date; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_user_purchases_date ON public.user_purchases USING btree (purchased_at DESC);


--
-- Name: idx_user_purchases_product; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_user_purchases_product ON public.user_purchases USING btree (product_id);


--
-- Name: idx_user_purchases_user; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_user_purchases_user ON public.user_purchases USING btree (user_id);


--
-- Name: idx_user_subscriptions_expires_at; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_user_subscriptions_expires_at ON public.user_subscriptions USING btree (expires_at);


--
-- Name: idx_user_subscriptions_user_id; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_user_subscriptions_user_id ON public.user_subscriptions USING btree (user_id);


--
-- Name: idx_users_balance; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_users_balance ON public.users USING btree (balance);


--
-- Name: idx_users_created_at; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_users_created_at ON public.users USING btree (created_at DESC);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: idx_users_referral_code; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_users_referral_code ON public.users USING btree (referral_code);


--
-- Name: idx_users_username; Type: INDEX; Schema: public; Owner: pixelcraft_user
--

CREATE INDEX idx_users_username ON public.users USING btree (username);


--
-- Name: products update_products_updated_at; Type: TRIGGER; Schema: public; Owner: pixelcraft_user
--

CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON public.products FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: subscriptions update_subscriptions_updated_at; Type: TRIGGER; Schema: public; Owner: pixelcraft_user
--

CREATE TRIGGER update_subscriptions_updated_at BEFORE UPDATE ON public.subscriptions FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: users update_users_updated_at; Type: TRIGGER; Schema: public; Owner: pixelcraft_user
--

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: discounts discounts_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.discounts
    ADD CONSTRAINT discounts_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: files files_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.files
    ADD CONSTRAINT files_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(id);


--
-- Name: library library_payment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.library
    ADD CONSTRAINT library_payment_id_fkey FOREIGN KEY (payment_id) REFERENCES public.payments(id);


--
-- Name: library library_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.library
    ADD CONSTRAINT library_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: library library_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.library
    ADD CONSTRAINT library_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: payments payments_subscription_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT payments_subscription_id_fkey FOREIGN KEY (subscription_id) REFERENCES public.subscriptions(id) ON DELETE SET NULL;


--
-- Name: payments payments_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT payments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: project_logs project_logs_subscription_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.project_logs
    ADD CONSTRAINT project_logs_subscription_id_fkey FOREIGN KEY (subscription_id) REFERENCES public.subscriptions(id);


--
-- Name: subscriptions subscriptions_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES public.plans(id);


--
-- Name: subscriptions subscriptions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: transactions transactions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: user_purchases user_purchases_payment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.user_purchases
    ADD CONSTRAINT user_purchases_payment_id_fkey FOREIGN KEY (payment_id) REFERENCES public.payments(id);


--
-- Name: user_purchases user_purchases_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.user_purchases
    ADD CONSTRAINT user_purchases_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE RESTRICT;


--
-- Name: user_purchases user_purchases_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.user_purchases
    ADD CONSTRAINT user_purchases_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_subscriptions user_subscriptions_payment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.user_subscriptions
    ADD CONSTRAINT user_subscriptions_payment_id_fkey FOREIGN KEY (payment_id) REFERENCES public.payments(id) ON DELETE SET NULL;


--
-- Name: user_subscriptions user_subscriptions_subscription_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.user_subscriptions
    ADD CONSTRAINT user_subscriptions_subscription_id_fkey FOREIGN KEY (subscription_id) REFERENCES public.subscriptions(id) ON DELETE CASCADE;


--
-- Name: user_subscriptions user_subscriptions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: pixelcraft_user
--

ALTER TABLE ONLY public.user_subscriptions
    ADD CONSTRAINT user_subscriptions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict wXuNeXnb6IN23NbXw7myZdFuQSwJmMpFTo7lgewyepthXjaaNX25DZDBeABLA3b

