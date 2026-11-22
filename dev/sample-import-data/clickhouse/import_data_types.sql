INSERT INTO test_db.data_types (
    id, col_int8, col_int16, col_int32, col_int64,
    col_float32, col_float64, col_decimal, col_decimal32, col_decimal64,
    col_date, col_datetime, col_string, col_fixedstring, col_boolean,
    col_enum8, col_enum16, col_nullable_int, col_nullable_string,
    col_lowcard, col_lowcard_nullable,
    col_array, col_array_string, col_map, col_tuple
) VALUES
(2, 10, 200, 30000, 4000000000, 3.5, 4.5, 456.78, 456.78901, 456.789, '2025-02-01', '2025-02-01 10:00:00', 'string_val_2', 'fixed_two1', 1, 'active', 'small', 10, 'import_a', 'cat_x', 'tag_x', [1, 2, 3], ['a', 'b'], {'k1': 10}, ('imp', 1, 1.1)),
(3, -5, -20, -3000, -400000000, -1.5, -2.5, -12.34, -12.34000, -12.340, '2025-02-02', '2025-02-02 11:00:00', 'string_val_3', 'fixed_thr1', 0, 'inactive', 'large', NULL, NULL, 'cat_y', NULL, [4, 5], ['c'], {'k2': 20}, ('imp2', 2, 2.2));