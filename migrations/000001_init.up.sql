CREATE TABLE restaurants (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE menu_items (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    restaurant_id BIGINT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    price BIGINT NOT NULL,
    is_available BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_menu_items_restaurant
        FOREIGN KEY (restaurant_id)
        REFERENCES restaurants(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_menu_items_price
        CHECK (price >= 0)
);

CREATE TABLE orders (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    restaurant_id BIGINT NOT NULL,
    status TEXT NOT NULL DEFAULT 'CREATED',
    total_price BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_orders_restaurant
        FOREIGN KEY (restaurant_id)
        REFERENCES restaurants(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_orders_status
        CHECK (
            status IN (
                'CREATED',
                'ACCEPTED',
                'PREPARING',
                'READY',
                'COMPLETED',
                'REJECTED'
            )
        ),

    CONSTRAINT chk_orders_total_price
        CHECK (total_price >= 0)
);

CREATE TABLE order_items (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_id BIGINT NOT NULL,
    menu_item_id BIGINT NOT NULL,
    name TEXT NOT NULL,
    unit_price BIGINT NOT NULL,
    quantity INTEGER NOT NULL,

    CONSTRAINT fk_order_items_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_order_items_menu_item
        FOREIGN KEY (menu_item_id)
        REFERENCES menu_items(id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_order_items_unit_price
        CHECK (unit_price >= 0),

    CONSTRAINT chk_order_items_quantity
        CHECK (quantity > 0)
);

CREATE INDEX idx_menu_items_restaurant_id
    ON menu_items (restaurant_id);

CREATE INDEX idx_orders_restaurant_status
    ON orders (restaurant_id, status);

CREATE INDEX idx_order_items_order_id
    ON order_items (order_id);