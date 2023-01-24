
# Create new sensitive_data_scanner_rule resource


resource "datadog_sensitive_data_scanner_rule" "foo" {
  description         = "UPDATE ME"
  excluded_attributes = ["username"]
  is_enabled          = "UPDATE ME"
  name                = "UPDATE ME"
  pattern             = "UPDATE ME"
  tags                = "UPDATE ME"
  text_replacement {
    number_of_chars    = "UPDATE ME"
    replacement_string = "UPDATE ME"
    type               = "UPDATE ME"
  }
}