-- +goose Up
-- Correct preview imports made before INDmoney's Total Value was identified
-- as current valuation. Preserve snapshot identity, dates, and original fields.
UPDATE investment_snapshots
SET data_json = json_set(data_json,
    '$.current_value', json_extract(data_json, '$.invested_value'),
    '$.invested_value', NULL,
    '$.holdings', json((SELECT json_group_array(json_set(value,
        '$.current_value', json_extract(value, '$.invested_value'),
        '$.invested_value', NULL,
        '$.closing_price', json_extract(value, '$.average_price'),
        '$.average_price', NULL
    )) FROM json_each(data_json, '$.holdings'))),
    '$.warnings', json_array('This statement provides current holding values in USD. Acquisition costs are not provided, so invested amount and returns are unavailable. No currency conversion is applied.')
)
WHERE json_extract(data_json, '$.parser_id') = 'indmoney_us_holdings_xls_v1'
  AND json_extract(data_json, '$.current_value') IS NULL
  AND json_extract(data_json, '$.invested_value') IS NOT NULL;

-- +goose Down
UPDATE investment_snapshots
SET data_json = json_set(data_json,
    '$.invested_value', json_extract(data_json, '$.current_value'),
    '$.current_value', NULL,
    '$.holdings', json((SELECT json_group_array(json_set(value,
        '$.invested_value', json_extract(value, '$.current_value'),
        '$.current_value', NULL,
        '$.average_price', json_extract(value, '$.closing_price'),
        '$.closing_price', NULL
    )) FROM json_each(data_json, '$.holdings'))),
    '$.warnings', json_array('This statement provides average acquisition prices and holding costs in USD. Market prices, current value, and returns are not provided. No currency conversion is applied.')
)
WHERE json_extract(data_json, '$.parser_id') = 'indmoney_us_holdings_xls_v1'
  AND json_extract(data_json, '$.invested_value') IS NULL
  AND json_extract(data_json, '$.current_value') IS NOT NULL;
