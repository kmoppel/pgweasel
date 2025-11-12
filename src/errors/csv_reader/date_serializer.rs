use chrono::{DateTime, FixedOffset, NaiveDateTime, TimeZone};
use serde::{self, Deserialize, Deserializer, Serializer};

const FORMAT: &str = "%Y-%m-%d %H:%M:%S%.3f %Z";

pub fn serialize<S>(date: &Option<DateTime<FixedOffset>>, serializer: S) -> Result<S::Ok, S::Error>
where
    S: Serializer,
{
    match date {
        Some(dt) => serializer.serialize_str(&dt.format(FORMAT).to_string()),
        None => serializer.serialize_none(),
    }
}

pub fn deserialize<'de, D>(deserializer: D) -> Result<Option<DateTime<FixedOffset>>, D::Error>
where
    D: Deserializer<'de>,
{
    let s: Option<String> = Option::deserialize(deserializer)?;
    if let Some(s) = s {
        // Attempt parsing using chrono’s parser with timezone name
        if let Ok(dt) = DateTime::parse_from_str(&s, FORMAT) {
            return Ok(Some(dt));
        }

        // Fallback: parse without timezone and assume EEST (+03:00)
        if let Ok(naive) =
            NaiveDateTime::parse_from_str(&s.replace(" EEST", ""), "%Y-%m-%d %H:%M:%S%.3f")
        {
            let offset = FixedOffset::east_opt(3 * 3600).unwrap();
            return Ok(Some(offset.from_local_datetime(&naive).unwrap()));
        }

        return Err(serde::de::Error::custom(format!(
            "Invalid datetime format: {}",
            s
        )));
    }
    Ok(None)
}
