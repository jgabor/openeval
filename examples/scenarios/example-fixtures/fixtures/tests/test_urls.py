import unittest

from maintainer_tools.urls import redact_credentials


class RedactCredentialsTests(unittest.TestCase):
    def test_redacts_credentials_and_preserves_other_parameters(self):
        url = "https://example.test/jobs?token=secret&limit=20&api_key=key"
        self.assertEqual(
            redact_credentials(url),
            "https://example.test/jobs?token=REDACTED&limit=20&api_key=REDACTED",
        )
        self.assertEqual(
            redact_credentials("https://example.test/?password=hunter2#status"),
            "https://example.test/?password=REDACTED#status",
        )


if __name__ == "__main__":
    unittest.main()
