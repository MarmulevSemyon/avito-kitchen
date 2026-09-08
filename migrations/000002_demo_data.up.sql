INSERT INTO restaurants (
    name,
    description,
    is_active
)
VALUES (
    'Demo Restaurant',
    'Demo restaurant for Avito Kitchen',
    TRUE
);

INSERT INTO menu_items (
    restaurant_id,
    name,
    description,
    price,
    is_available
)
VALUES
    (
        1,
        'Burger',
        'Beef burger',
        50000,
        TRUE
    ),
    (
        1,
        'Pizza',
        'Margherita pizza',
        70000,
        TRUE
    ),
    (
        1,
        'Cola',
        '330 ml',
        15000,
        TRUE
    );