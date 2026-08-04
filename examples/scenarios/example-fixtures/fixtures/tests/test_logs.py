import unittest

from maintainer_tools.logs import count_log_levels


class CountLogLevelsTests(unittest.TestCase):
    def test_counts_levels_case_insensitively_and_ignores_unknown_lines(self):
        lines = ["[info] started", "[ERROR] failed", "plain output", "[Info] done"]
        self.assertEqual(count_log_levels(lines), {"INFO": 2, "ERROR": 1})


if __name__ == "__main__":
    unittest.main()
