use chrono::{DateTime, FixedOffset};

use crate::Result;

const FORMAT: &str = "%Y-%m-%d %H:%M:%S%.3f %:z";

pub fn deserialize_helper(s: &str) -> Result<DateTime<FixedOffset>> {
    if let Ok(dt) = DateTime::parse_from_str(s, FORMAT) {
        return Ok(dt);
    }

    let s_replaced = s
        .replace("UTC", "+00:00")
        .replace("PST", "-08:00")
        .replace("PDT", "-07:00")
        .replace("EEST", "+03:00")
        .replace("EST", "-05:00")
        .replace("EDT", "-04:00")
        .replace("CST", "-06:00")
        .replace("CDT", "-05:00")
        .replace("MST", "-07:00")
        .replace("MDT", "-06:00")
        .replace("EET", "+02:00");

    DateTime::parse_from_str(&s_replaced, FORMAT)
        .map_err(|err| format!("Invalid datetime format: {} Error: {}", s, err).into())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_deserialize_with_eest() {
        let original_str = "2023-10-05 14:30:15.123 EEST";
        let dt = deserialize_helper(original_str).unwrap();
        let serialized_str = dt.format(FORMAT).to_string();
        assert_eq!(serialized_str, "2023-10-05 14:30:15.123 +03:00".to_string());
    }

    #[test]
    fn test_deserialize_with_utc() {
        let original_str = "2025-11-11 06:25:53.178 UTC";
        let dt = deserialize_helper(original_str).unwrap();
        let serialized_str = dt.format(FORMAT).to_string();
        assert_eq!(serialized_str, "2025-11-11 06:25:53.178 +00:00".to_string());
    }
}
