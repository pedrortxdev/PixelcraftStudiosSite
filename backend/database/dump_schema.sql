--
-- PostgreSQL database dump
--

\restrict rKecbVC9sY6ezeySciA323QcAO2gVR1xtc0XZJYhyqjdpIdBWUqa4MQuWGs8g7l

-- Dumped from database version 18.0
-- Dumped by pg_dump version 18.0

SET search_path = public;
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

--
-- Name: pixelcraft; Type: DATABASE; Schema: -; Owner: postgres
--




\unrestrict rKecbVC9sY6ezeySciA323QcAO2gVR1xtc0XZJYhyqjdpIdBWUqa4MQuWGs8g7l
\connect pixelcraft
\restrict rKecbVC9sY6ezeySciA323QcAO2gVR1xtc0XZJYhyqjdpIdBWUqa4MQuWGs8g7l

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
-- Name: discount_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.discount_type AS ENUM (
    'PERCENTAGE',
    'FIXED_AMOUNT'
);



--
-- Name: payment_status; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.payment_status AS ENUM (
    'PENDING',
    'COMPLETED',
    'FAILED',
    'REFUNDED'
);



--
-- Name: product_type; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.product_type AS ENUM (
    'PLUGIN',
    'MOD',
    'MAP',
    'TEXTUREPACK',
    'SERVER_TEMPLATE'
);



--
-- Name: subscription_status; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.subscription_status AS ENUM (
    'ACTIVE',
    'CANCELED',
    'PAST_DUE'
);



--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;



SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: discounts; Type: TABLE; Schema: public; Owner: postgres
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



--
-- Name: TABLE discounts; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.discounts IS 'Cupons de desconto e cÃ³digos de indicaÃ§Ã£o';


--
-- Name: COLUMN discounts.is_referral; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.discounts.is_referral IS 'TRUE se for cÃ³digo de indicaÃ§Ã£o';


--
-- Name: COLUMN discounts.created_by_user_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.discounts.created_by_user_id IS 'UsuÃ¡rio que gerou o cÃ³digo de referral';


--
-- Name: library; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.library (
    id uuid NOT NULL,
    user_id uuid NOT NULL,
    product_id uuid NOT NULL,
    payment_id uuid NOT NULL,
    purchased_at timestamp with time zone DEFAULT now()
);



--
-- Name: payments; Type: TABLE; Schema: public; Owner: postgres
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



--
-- Name: TABLE payments; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.payments IS 'Registro completo de todas as transaÃ§Ãµes financeiras';


--
-- Name: COLUMN payments.subscription_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.payments.subscription_id IS 'NULL para compras avulsas, UUID para renovaÃ§Ãµes de assinatura';


--
-- Name: COLUMN payments.payment_gateway_id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.payments.payment_gateway_id IS 'ID externo do gateway de pagamento';


--
-- Name: COLUMN payments.payment_method; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.payments.payment_method IS 'MÃ©todo usado: BALANCE, STRIPE, MERCADOPAGO, etc.';


--
-- Name: products; Type: TABLE; Schema: public; Owner: postgres
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



--
-- Name: TABLE products; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.products IS 'CatÃ¡logo de produtos digitais (plugins, mods, maps, etc.)';


--
-- Name: COLUMN products.download_url_encrypted; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.products.download_url_encrypted IS 'URL de download criptografada com pgp_sym_encrypt';


--
-- Name: COLUMN products.is_exclusive; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.products.is_exclusive IS 'Produto de ediÃ§Ã£o limitada (venda Ãºnica)';


--
-- Name: COLUMN products.stock_quantity; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.products.stock_quantity IS 'Quantidade em estoque (NULL = ilimitado)';


--
-- Name: subscriptions; Type: TABLE; Schema: public; Owner: postgres
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
    CONSTRAINT subscriptions_price_per_month_check CHECK ((price_per_month >= (0)::numeric))
);



--
-- Name: TABLE subscriptions; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.subscriptions IS 'Assinaturas mensais de planos de desenvolvimento';


--
-- Name: COLUMN subscriptions.next_billing_date; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subscriptions.next_billing_date IS 'Data da prÃ³xima cobranÃ§a automÃ¡tica';


--
-- Name: COLUMN subscriptions.plan_metadata; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.subscriptions.plan_metadata IS 'Features e configuraÃ§Ãµes do plano em JSON';


--
-- Name: test; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.test (
    id integer NOT NULL,
    name character varying(50)
);



--
-- Name: test_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.test_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



--
-- Name: test_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.test_id_seq OWNED BY public.test.id;


--
-- Name: user_purchases; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_purchases (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    user_id uuid NOT NULL,
    product_id uuid NOT NULL,
    purchase_price numeric(10,2) NOT NULL,
    purchased_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT user_purchases_purchase_price_check CHECK ((purchase_price >= (0)::numeric))
);



--
-- Name: TABLE user_purchases; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.user_purchases IS 'Registro de produtos comprados por cada usuÃ¡rio';


--
-- Name: COLUMN user_purchases.purchase_price; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.user_purchases.purchase_price IS 'PreÃ§o pago no momento da compra (histÃ³rico)';


--
-- Name: user_subscriptions; Type: TABLE; Schema: public; Owner: postgres
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



--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
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



--
-- Name: TABLE users; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON TABLE public.users IS 'Tabela principal de usuÃ¡rios com autenticaÃ§Ã£o e dados de perfil';


--
-- Name: COLUMN users.id; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.id IS 'Identificador Ãºnico UUID do usuÃ¡rio';


--
-- Name: COLUMN users.username; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.username IS 'Nome de usuÃ¡rio Ãºnico (login)';


--
-- Name: COLUMN users.email; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.email IS 'Email Ãºnico do usuÃ¡rio';


--
-- Name: COLUMN users.password_hash; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.password_hash IS 'Hash bcrypt da senha';


--
-- Name: COLUMN users.full_name; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.full_name IS 'Nome completo do usuÃ¡rio (opcional)';


--
-- Name: COLUMN users.discord_handle; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.discord_handle IS 'Handle do Discord (opcional)';


--
-- Name: COLUMN users.whatsapp_phone; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.whatsapp_phone IS 'NÃºmero do WhatsApp (opcional)';


--
-- Name: COLUMN users.cpf_encrypted; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.cpf_encrypted IS 'CPF criptografado com pgp_sym_encrypt (para faturamento)';


--
-- Name: COLUMN users.created_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.created_at IS 'Timestamp de criaÃ§Ã£o da conta';


--
-- Name: COLUMN users.updated_at; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.updated_at IS 'Timestamp da Ãºltima atualizaÃ§Ã£o';


--
-- Name: COLUMN users.balance; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.balance IS 'Saldo disponÃ­vel do usuÃ¡rio para compras';


--
-- Name: COLUMN users.referral_code; Type: COMMENT; Schema: public; Owner: postgres
--

COMMENT ON COLUMN public.users.referral_code IS 'CÃ³digo de referÃªncia Ãºnico do usuÃ¡rio para indicar amigos';


--
-- Name: test id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.test ALTER COLUMN id SET DEFAULT nextval('public.test_id_seq'::regclass);


--
-- Name: discounts discounts_code_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.discounts
    ADD CONSTRAINT discounts_code_key UNIQUE (code);


--
-- Name: discounts discounts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.discounts
    ADD CONSTRAINT discounts_pkey PRIMARY KEY (id);


--
-- Name: library library_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.library
    ADD CONSTRAINT library_pkey PRIMARY KEY (id);


--
-- Name: payments payments_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT payments_pkey PRIMARY KEY (id);


--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);


--
-- Name: subscriptions subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_pkey PRIMARY KEY (id);


--
-- Name: test test_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.test
    ADD CONSTRAINT test_pkey PRIMARY KEY (id);


--
-- Name: library unique_user_product; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.library
    ADD CONSTRAINT unique_user_product UNIQUE (user_id, product_id);


--
-- Name: user_purchases user_purchases_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_purchases
    ADD CONSTRAINT user_purchases_pkey PRIMARY KEY (id);


--
-- Name: user_purchases user_purchases_user_id_product_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_purchases
    ADD CONSTRAINT user_purchases_user_id_product_id_key UNIQUE (user_id, product_id);


--
-- Name: user_subscriptions user_subscriptions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_subscriptions
    ADD CONSTRAINT user_subscriptions_pkey PRIMARY KEY (id);


--
-- Name: user_subscriptions user_subscriptions_user_id_subscription_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_subscriptions
    ADD CONSTRAINT user_subscriptions_user_id_subscription_id_key UNIQUE (user_id, subscription_id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: users users_referral_code_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_referral_code_key UNIQUE (referral_code);


--
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- Name: idx_discounts_active; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_discounts_active ON public.discounts USING btree (is_active);


--
-- Name: idx_discounts_code; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_discounts_code ON public.discounts USING btree (code);


--
-- Name: idx_discounts_creator; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_discounts_creator ON public.discounts USING btree (created_by_user_id);


--
-- Name: idx_library_product; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_library_product ON public.library USING btree (product_id);


--
-- Name: idx_library_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_library_user ON public.library USING btree (user_id);


--
-- Name: idx_payments_created_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payments_created_at ON public.payments USING btree (created_at DESC);


--
-- Name: idx_payments_gateway_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payments_gateway_id ON public.payments USING btree (payment_gateway_id);


--
-- Name: idx_payments_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payments_status ON public.payments USING btree (status);


--
-- Name: idx_payments_subscription; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payments_subscription ON public.payments USING btree (subscription_id);


--
-- Name: idx_payments_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_payments_user ON public.payments USING btree (user_id);


--
-- Name: idx_products_active; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_products_active ON public.products USING btree (is_active);


--
-- Name: idx_products_price; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_products_price ON public.products USING btree (price);


--
-- Name: idx_products_type; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_products_type ON public.products USING btree (type);


--
-- Name: idx_subscriptions_billing_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_subscriptions_billing_date ON public.subscriptions USING btree (next_billing_date);


--
-- Name: idx_subscriptions_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_subscriptions_status ON public.subscriptions USING btree (status);


--
-- Name: idx_subscriptions_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_subscriptions_user ON public.subscriptions USING btree (user_id);


--
-- Name: idx_user_purchases_date; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_purchases_date ON public.user_purchases USING btree (purchased_at DESC);


--
-- Name: idx_user_purchases_product; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_purchases_product ON public.user_purchases USING btree (product_id);


--
-- Name: idx_user_purchases_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_purchases_user ON public.user_purchases USING btree (user_id);


--
-- Name: idx_user_subscriptions_expires_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_subscriptions_expires_at ON public.user_subscriptions USING btree (expires_at);


--
-- Name: idx_user_subscriptions_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_user_subscriptions_user_id ON public.user_subscriptions USING btree (user_id);


--
-- Name: idx_users_balance; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_balance ON public.users USING btree (balance);


--
-- Name: idx_users_created_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_created_at ON public.users USING btree (created_at DESC);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: idx_users_referral_code; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_referral_code ON public.users USING btree (referral_code);


--
-- Name: idx_users_username; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_username ON public.users USING btree (username);


--
-- Name: products update_products_updated_at; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON public.products FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: subscriptions update_subscriptions_updated_at; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_subscriptions_updated_at BEFORE UPDATE ON public.subscriptions FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: users update_users_updated_at; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: discounts discounts_created_by_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.discounts
    ADD CONSTRAINT discounts_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON DELETE SET NULL;


--
-- Name: library library_payment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.library
    ADD CONSTRAINT library_payment_id_fkey FOREIGN KEY (payment_id) REFERENCES public.payments(id);


--
-- Name: library library_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.library
    ADD CONSTRAINT library_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE CASCADE;


--
-- Name: library library_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.library
    ADD CONSTRAINT library_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: payments payments_subscription_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT payments_subscription_id_fkey FOREIGN KEY (subscription_id) REFERENCES public.subscriptions(id) ON DELETE SET NULL;


--
-- Name: payments payments_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT payments_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: subscriptions subscriptions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.subscriptions
    ADD CONSTRAINT subscriptions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_purchases user_purchases_product_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_purchases
    ADD CONSTRAINT user_purchases_product_id_fkey FOREIGN KEY (product_id) REFERENCES public.products(id) ON DELETE RESTRICT;


--
-- Name: user_purchases user_purchases_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_purchases
    ADD CONSTRAINT user_purchases_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: user_subscriptions user_subscriptions_payment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

---
--- Admin Analytics Snapshot
---

ALTER TABLE ONLY public.user_subscriptions
    ADD CONSTRAINT user_subscriptions_payment_id_fkey FOREIGN KEY (payment_id) REFERENCES public.payments(id) ON DELETE SET NULL;

CREATE TABLE admin_analytics_snapshot (
    id SERIAL PRIMARY KEY,
    total_revenue DECIMAL(15, 2) DEFAULT 0,
    total_users INT DEFAULT 0,
    active_products INT DEFAULT 0,
    total_sales INT DEFAULT 0,
    revenue_growth_pct DECIMAL(5, 2) DEFAULT 0,
    users_growth_pct DECIMAL(5, 2) DEFAULT 0,
    products_status VARCHAR(50) DEFAULT 'Estável',
    sales_growth_pct DECIMAL(5, 2) DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insere um dado fake só pra API não retornar vazio no teste
INSERT INTO admin_analytics_snapshot 
(total_revenue, total_users, active_products, total_sales, revenue_growth_pct, users_growth_pct, sales_growth_pct)
VALUES (45678.90, 1234, 89, 456, 23.5, 18.2, 15.8);

-- Limpa se tiver lixo e insere o inicializador zerado
TRUNCATE TABLE admin_analytics_snapshot;
INSERT INTO admin_analytics_snapshot (id, total_revenue, total_users, active_products, total_sales)
VALUES (1, 0, 0, 0, 0);

CREATE OR REPLACE FUNCTION update_user_stats() RETURNS TRIGGER AS $$
BEGIN
    -- Se for INSERT (Novo usuário)
    IF (TG_OP = 'INSERT') THEN
        UPDATE admin_analytics_snapshot
        SET total_users = total_users + 1,
            last_updated = NOW()
        WHERE id = 1;
    -- Se for DELETE (Caso você permita deletar usuário)
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE admin_analytics_snapshot
        SET total_users = total_users - 1,
            last_updated = NOW()
        WHERE id = 1;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- O Gatilho
CREATE TRIGGER trg_update_user_stats
AFTER INSERT OR DELETE ON users
FOR EACH ROW EXECUTE FUNCTION update_user_stats();

CREATE OR REPLACE FUNCTION update_sales_stats() RETURNS TRIGGER AS $$
BEGIN
    -- Só conta se o pagamento for confirmado
    IF (NEW.status = 'COMPLETED') THEN
        UPDATE admin_analytics_snapshot
        SET 
            total_sales = total_sales + 1,
            total_revenue = total_revenue + NEW.final_amount,
            last_updated = NOW()
        WHERE id = 1;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- O Gatilho
CREATE TRIGGER trg_update_sales_stats
AFTER INSERT OR UPDATE ON payments -- ou 'orders', veja qual tabela registra o $$$
FOR EACH ROW EXECUTE FUNCTION update_sales_stats();

CREATE OR REPLACE FUNCTION update_product_stats() RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        UPDATE admin_analytics_snapshot SET active_products = active_products + 1 WHERE id = 1;
    ELSIF (TG_OP = 'DELETE') THEN
        UPDATE admin_analytics_snapshot SET active_products = active_products - 1 WHERE id = 1;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_product_stats
AFTER INSERT OR DELETE ON products
FOR EACH ROW EXECUTE FUNCTION update_product_stats();

ALTER TABLE admin_analytics_snapshot 
RENAME COLUMN updated_at TO last_updated;
--
-- Name: user_subscriptions user_subscriptions_subscription_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_subscriptions
    ADD CONSTRAINT user_subscriptions_subscription_id_fkey FOREIGN KEY (subscription_id) REFERENCES public.subscriptions(id) ON DELETE CASCADE;


--
-- Name: user_subscriptions user_subscriptions_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_subscriptions
    ADD CONSTRAINT user_subscriptions_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict rKecbVC9sY6ezeySciA323QcAO2gVR1xtc0XZJYhyqjdpIdBWUqa4MQuWGs8g7l

