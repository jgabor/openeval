import unittest

from maintainer_tools.accounts import normalize_account_name


class NormalizeAccountNameTests(unittest.TestCase):
    def test_normalizes_surrounding_and_repeated_whitespace(self):
        self.assertEqual(normalize_account_name("  Release   Engineering  "), "release-engineering")
        self.assertEqual(normalize_account_name("Quality\tAssurance"), "quality-assurance")


if __name__ == "__main__":
    unittest.main()
