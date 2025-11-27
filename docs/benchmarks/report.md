# QH Protocol Benchmark Report

**Generated:** November 27, 2025 at 11:37:55  
**Commit:** `94ba292`  
**Go Version:** go1.25.1  
**Platform:** Darwin/x86_64  

**110 Test Cases** (10 Edge Cases + 100 Real Traffic)

## Table of Contents

1. [Edge Case Analysis](#edge-case-analysis) (10 test cases)
2. [Real HTTP Traffic Analysis](#real-http-traffic-analysis) (100 test cases)
3. [Combined Results](#combined-results) (110 test cases)

## Edge Case Analysis

### Summary

- **10** test cases, manually selected
- QH total: **4756 B**
- HTTP/1.1 total: **7044 B** (48.1% larger)
- HTTP/2 total: **4442 B** (6.6% smaller)
- HTTP/3 total: **4094 B** (13.9% smaller)

### Performance Bounds

**vs HTTP/1.1:**

- **Best case**: 92.0% smaller - QH: 49 B vs HTTP/1.1: 610 B (Edge Case 1: QH Best Case - All Static Table Complete Pairs)
- **Worst case**: 11.6% smaller - QH: 1062 B vs HTTP/1.1: 1201 B (Edge Case 8: Verbose Headers)

**vs HTTP/2:**

- **Best case**: 85.5% smaller - QH: 49 B vs HTTP/2: 337 B (Edge Case 1: QH Best Case - All Static Table Complete Pairs)
- **Worst case**: 29.2% larger - QH: 1062 B vs HTTP/2: 822 B (Edge Case 8: Verbose Headers)

**vs HTTP/3:**

- **Best case**: 82.1% smaller - QH: 49 B vs HTTP/3: 274 B (Edge Case 1: QH Best Case - All Static Table Complete Pairs)
- **Worst case**: 35.5% larger - QH: 858 B vs HTTP/3: 633 B (Edge Case 9: Large Single Header - Security Policy)

<details>
<summary><strong>Click to expand detailed test case results</strong></summary>

| Test Case | QH (bytes) | HTTP/1 | HTTP/2 | HTTP/3 | QH/H1 | QH/H2 | QH/H3 |
|-----------|----------:|-------:|-------:|-------:|------:|------:|------:|
| Edge Case 1: QH Best Case - All Static Table Co... | 49 | 610 | 337 | 274 | 8.0% | 14.5% | 17.9% |
| Edge Case 2: QH Worst Case - All Custom Headers | 824 | 947 | 640 | 632 | 87.0% | 128.8% | 130.4% |
| Edge Case 3: HTTP/1.1 Best Case - Minimal Headers | 34 | 190 | 71 | 49 | 17.9% | 47.9% | 69.4% |
| Edge Case 4: QH Format 2 - Static Names, Custom... | 637 | 821 | 520 | 517 | 77.6% | 122.5% | 123.2% |
| Edge Case 5: QH Mixed Formats | 246 | 541 | 297 | 266 | 45.5% | 82.8% | 92.5% |
| Edge Case 6: Many Headers | 900 | 1217 | 791 | 754 | 74.0% | 113.8% | 119.4% |
| Edge Case 7: Minimal Response - 204 No Content | 53 | 147 | 65 | 50 | 36.1% | 81.5% | 106.0% |
| Edge Case 8: Verbose Headers | 1062 | 1201 | 822 | 801 | 88.4% | 129.2% | 132.6% |
| Edge Case 9: Large Single Header - Security Policy | 858 | 1000 | 676 | 633 | 85.8% | 126.9% | 135.5% |
| Edge Case 10: Empty Values - Present But Empty ... | 93 | 370 | 223 | 118 | 25.1% | 41.7% | 78.8% |

</details>

### Request/Response Breakdown

**Request Headers:**

- QH avg: **279 B** (baseline)
- HTTP/1 avg: **413 B** (47.9% larger)
- HTTP/2 avg: **261 B** (6.5% smaller)
- HTTP/3 avg: **245 B** (12.0% smaller)

**Response Headers:**

- QH avg: **197 B** (baseline)
- HTTP/1 avg: **292 B** (48.3% larger)
- HTTP/2 avg: **183 B** (6.8% smaller)
- HTTP/3 avg: **164 B** (16.6% smaller)

**Total Headers (Request + Response):**

- QH avg: **476 B** (baseline)
- HTTP/1 avg: **704 B** (48.1% larger)
- HTTP/2 avg: **444 B** (6.6% smaller)
- HTTP/3 avg: **409 B** (13.9% smaller)

## Real HTTP Traffic Analysis

### Summary

- **100** test cases, collected from actual internet traffic
- QH total: **140195 B**
- HTTP/1.1 total: **194200 B** (38.5% larger)
- HTTP/2 total: **129472 B** (7.6% smaller)
- HTTP/3 total: **123813 B** (11.7% smaller)

### Performance Bounds

**vs HTTP/1.1:**

- **Best case**: 51.5% smaller - QH: 480 B vs HTTP/1.1: 990 B (Request 93: OPTIONS /models)
- **Worst case**: 7.9% smaller - QH: 4196 B vs HTTP/1.1: 4557 B (Request 1: GET /manifest.json)

**vs HTTP/2:**

- **Best case**: 23.6% smaller - QH: 480 B vs HTTP/2: 628 B (Request 93: OPTIONS /models)
- **Worst case**: 34.0% larger - QH: 926 B vs HTTP/2: 691 B (Request 95: POST /t/perf?aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa)

**vs HTTP/3:**

- **Best case**: 18.2% smaller - QH: 968 B vs HTTP/3: 1184 B (Request 31: GET /)
- **Worst case**: 36.9% larger - QH: 4196 B vs HTTP/3: 3066 B (Request 1: GET /manifest.json)

### Breakdown by Size Category

| Category | Count | QH Avg | H1 Avg | H2 Avg | H3 Avg | QH/H1 | QH/H2 | QH/H3 |
|----------|------:|-------:|-------:|-------:|-------:|------:|------:|------:|
| Tiny (<1KB) | 66 | 697 B | 1.2 KB | 760 B | 715 B | 58.5% | 91.7% | 97.4% |
| Small (1-10KB) | 34 | 2.7 KB | 3.3 KB | 2.3 KB | 2.2 KB | 81.5% | 118.8% | 123.0% |

<details>
<summary><strong>Click to expand detailed test case results</strong></summary>

| Test Case | QH (bytes) | HTTP/1 | HTTP/2 | HTTP/3 | QH/H1 | QH/H2 | QH/H3 |
|-----------|----------:|-------:|-------:|-------:|------:|------:|------:|
| Request 1: GET /manifest.json | 4196 | 4557 | 3143 | 3066 | 92.1% | 133.5% | 136.9% |
| Request 2: POST /github-copilot/chat/token | 4335 | 5007 | 3428 | 3339 | 86.6% | 126.5% | 129.8% |
| Request 3: GET /_next/data/qcPKpvHcKq6nT9zVikei... | 503 | 872 | 563 | 543 | 57.7% | 89.3% | 92.6% |
| Request 4: GET /_next/data/qcPKpvHcKq6nT9zVikei... | 514 | 885 | 572 | 552 | 58.1% | 89.9% | 93.1% |
| Request 5: GET /_next/data/qcPKpvHcKq6nT9zVikei... | 537 | 908 | 588 | 568 | 59.1% | 91.3% | 94.5% |
| Request 6: GET /_next/data/qcPKpvHcKq6nT9zVikei... | 493 | 862 | 557 | 537 | 57.2% | 88.5% | 91.8% |
| Request 7: GET /complete/search?q=astro%20docs&... | 1283 | 1999 | 1378 | 1291 | 64.2% | 93.1% | 99.4% |
| Request 8: GET /github-copilot/docs/docsets | 4391 | 4974 | 3439 | 3360 | 88.3% | 127.7% | 130.7% |
| Request 9: GET /complete/search?q=astro%aaaaaaa... | 1278 | 1994 | 1359 | 1272 | 64.1% | 94.0% | 100.5% |
| Request 10: POST /youtubei/v1/account/get_setti... | 1001 | 1565 | 1075 | 998 | 64.0% | 93.1% | 100.3% |
| Request 11: POST /youtubei/v1/player?prettyPrin... | 964 | 1533 | 1054 | 977 | 62.9% | 91.5% | 98.7% |
| Request 12: GET /_next/data/qcPKpvHcKq6nT9zVike... | 520 | 891 | 575 | 555 | 58.4% | 90.4% | 93.7% |
| Request 13: GET /models | 523 | 986 | 608 | 543 | 53.0% | 86.0% | 96.3% |
| Request 14: POST /$rpc/google.internal.waa.v1.W... | 703 | 1385 | 869 | 769 | 50.8% | 80.9% | 91.4% |
| Request 15: GET /_next/data/qcPKpvHcKq6nT9zVike... | 508 | 879 | 568 | 548 | 57.8% | 89.4% | 92.7% |
| Request 16: GET /github-copilot/chat | 4339 | 4999 | 3425 | 3343 | 86.8% | 126.7% | 129.8% |
| Request 17: GET /manifest.json | 4234 | 4694 | 3235 | 3163 | 90.2% | 130.9% | 133.9% |
| Request 18: POST /youtubei/v1/notification_regi... | 636 | 1066 | 686 | 608 | 59.7% | 92.7% | 104.6% |
| Request 19: GET /github-copilot/chat | 4361 | 5021 | 3442 | 3360 | 86.9% | 126.7% | 129.8% |
| Request 20: GET /sw.js_data | 1699 | 2310 | 1565 | 1478 | 73.5% | 108.6% | 115.0% |
| Request 21: HEAD /_next/data/qcPKpvHcKq6nT9zVik... | 496 | 867 | 564 | 551 | 57.2% | 87.9% | 90.0% |
| Request 22: GET /complete/search?q=aaaaaaaaaaaa... | 1271 | 1985 | 1354 | 1267 | 64.0% | 93.9% | 100.3% |
| Request 23: GET /manifest.webmanifest | 3125 | 3735 | 2558 | 2476 | 83.7% | 122.2% | 126.2% |
| Request 24: POST /youtubei/v1/next?prettyPrint=... | 981 | 1550 | 1067 | 990 | 63.3% | 91.9% | 99.1% |
| Request 25: GET /example-user/theme-hub/branch-... | 4339 | 4984 | 3419 | 3337 | 87.1% | 126.9% | 130.0% |
| Request 26: GET /complete/search?qaaaaaaaaaaaaa... | 1274 | 1988 | 1356 | 1269 | 64.1% | 94.0% | 100.4% |
| Request 27: POST /youtubei/v1/account/get_setti... | 982 | 1546 | 1063 | 986 | 63.5% | 92.4% | 99.6% |
| Request 28: GET /_next/data/qcPKpvHcKq6nT9zVike... | 502 | 873 | 562 | 542 | 57.5% | 89.3% | 92.6% |
| Request 29: GET /complete/search?aaaaaaaaaaaaaa... | 1262 | 2000 | 1363 | 1275 | 63.1% | 92.6% | 99.0% |
| Request 30: GET /youtube/img/lottie/subscribe_a... | 947 | 1502 | 1006 | 930 | 63.0% | 94.1% | 101.8% |
| Request 31: GET / | 968 | 1743 | 1227 | 1184 | 55.5% | 78.9% | 81.8% |
| Request 32: GET / | 4310 | 5024 | 3451 | 3341 | 85.8% | 124.9% | 129.0% |
| Request 33: GET /example-user/theme-hub/raw/mai... | 4314 | 4896 | 3360 | 3264 | 88.1% | 128.4% | 132.2% |
| Request 34: GET / | 4439 | 5035 | 3465 | 3343 | 88.2% | 128.1% | 132.8% |
| Request 35: GET /RotateCookiesPage?og_pid=538&r... | 1753 | 2692 | 1845 | 1737 | 65.1% | 95.0% | 100.9% |
| Request 36: GET /example-user/theme-hub | 4205 | 4925 | 3387 | 3277 | 85.4% | 124.2% | 128.3% |
| Request 37: GET /dashboard/my_top_repositories?... | 4506 | 5099 | 3505 | 3415 | 88.4% | 128.6% | 131.9% |
| Request 38: GET /github-copilot/chat?skip_ancho... | 4485 | 5078 | 3491 | 3401 | 88.3% | 128.5% | 131.9% |
| Request 39: GET /RotateCookiesPage?og_pid=1&rot... | 1751 | 2709 | 1857 | 1749 | 64.6% | 94.3% | 100.1% |
| Request 40: GET /api/stats/watchtime?ns=aaaaaaa... | 2173 | 2770 | 1890 | 1836 | 78.4% | 115.0% | 118.4% |
| Request 41: GET /assets/chunk-22809-58689ee661a... | 641 | 1237 | 754 | 705 | 51.8% | 85.0% | 90.9% |
| Request 42: GET /assets/chunk-72552-509f20f59e5... | 641 | 1237 | 751 | 702 | 51.8% | 85.4% | 91.3% |
| Request 43: GET /assets/chunk-64508-43a0bc75388... | 636 | 1108 | 675 | 629 | 57.4% | 94.2% | 101.1% |
| Request 44: GET /assets/chunk-89556-84a32902ac4... | 641 | 1237 | 753 | 704 | 51.8% | 85.1% | 91.1% |
| Request 45: GET /assets/codespaces-9f0a42ea762f.js | 640 | 1236 | 752 | 703 | 51.8% | 85.1% | 91.0% |
| Request 46: GET /assets/19718-676a65610616.js | 630 | 1102 | 672 | 626 | 57.2% | 93.8% | 100.6% |
| Request 47: GET /dist/s.862a716c937860ab008e.js | 2687 | 3107 | 2179 | 2110 | 86.5% | 123.3% | 127.3% |
| Request 48: GET /assets/34031-80252173b2e1.js | 636 | 1231 | 750 | 701 | 51.7% | 84.8% | 90.7% |
| Request 49: GET /s/desktop/b8106bc5/jsbin/www-i... | 607 | 976 | 632 | 566 | 62.2% | 96.0% | 107.2% |
| Request 50: GET /assets/51519-dc0d4e14166a.js | 635 | 1231 | 749 | 700 | 51.6% | 84.8% | 90.7% |
| Request 51: GET /_astro/page.DH-1p6s5.js | 526 | 1004 | 615 | 583 | 52.4% | 85.5% | 90.2% |
| Request 52: GET /assets/chunk-47657-9d37709c927... | 642 | 1237 | 754 | 705 | 51.9% | 85.1% | 91.1% |
| Request 53: GET /assets/copilot-chat-43e624a2f9... | 637 | 1107 | 676 | 630 | 57.5% | 94.2% | 101.1% |
| Request 54: GET /assets/octicons-react-2694eb47... | 646 | 1241 | 756 | 707 | 52.1% | 85.4% | 91.4% |
| Request 55: GET /assets/chunk-69458-4ad6513c42c... | 637 | 1109 | 676 | 630 | 57.4% | 94.2% | 101.1% |
| Request 56: GET /assets/chunk-79012-ad941ce0587... | 642 | 1238 | 754 | 705 | 51.9% | 85.1% | 91.1% |
| Request 57: GET /assets/72238-6d6ea226b976.js | 631 | 1103 | 672 | 626 | 57.2% | 93.9% | 100.8% |
| Request 58: GET /assets/34983-556f24fe0bd3.js | 634 | 1230 | 749 | 700 | 51.5% | 84.6% | 90.6% |
| Request 59: GET /_next/static/qcPKpvHcKq6nT9zVi... | 597 | 981 | 590 | 577 | 60.9% | 101.2% | 103.5% |
| Request 60: GET /_next/static/chunks/framework-... | 606 | 988 | 593 | 580 | 61.3% | 102.2% | 104.5% |
| Request 61: GET /assets/72568-d9b14327a489.js | 636 | 1232 | 751 | 702 | 51.6% | 84.7% | 90.6% |
| Request 62: GET /dist/util/u.8ec21e2c9eaa1212ed... | 2679 | 3288 | 2277 | 2191 | 81.5% | 117.7% | 122.3% |
| Request 63: GET /assets/87670-dfd94491d74f.js | 636 | 1232 | 750 | 701 | 51.6% | 84.8% | 90.7% |
| Request 64: GET /xjs/_/js/k=aaaaaaaaaaaaaaaaaaa... | 1902 | 2620 | 1741 | 1691 | 72.6% | 109.2% | 112.5% |
| Request 65: GET /assets/chunk-32300-b7c1a896bbb... | 642 | 1238 | 754 | 705 | 51.9% | 85.1% | 91.1% |
| Request 66: GET /assets/github-490b3f68bf33.css | 640 | 1053 | 656 | 633 | 60.8% | 97.6% | 101.1% |
| Request 67: GET /css2?family=Roboto:wght@300;40... | 1041 | 1801 | 1173 | 1092 | 57.8% | 88.7% | 95.3% |
| Request 68: GET /assets/74667.0a095c2859374624b... | 685 | 1263 | 774 | 739 | 54.2% | 88.5% | 92.7% |
| Request 69: GET /assets/code-9c9b8dc61e74.css | 639 | 1053 | 653 | 630 | 60.7% | 97.9% | 101.4% |
| Request 70: GET /assets/92017.192023d95dd142fdb... | 686 | 1264 | 775 | 740 | 54.3% | 88.5% | 92.7% |
| Request 71: GET /assets/44555.00b3eac1a85501f49... | 684 | 1262 | 773 | 738 | 54.2% | 88.5% | 92.7% |
| Request 72: GET /assets/copilot-chat.0a095c2859... | 654 | 1112 | 678 | 642 | 58.8% | 96.5% | 101.9% |
| Request 73: GET /assets/github-490b3f68bf33.css | 640 | 1053 | 656 | 633 | 60.8% | 97.6% | 101.1% |
| Request 74: GET /assets/74667.0a095c2859374624b... | 647 | 1105 | 673 | 637 | 58.6% | 96.1% | 101.6% |
| Request 75: GET /s/desktop/b8106bc5/cssbin/www-... | 983 | 1504 | 1033 | 961 | 65.4% | 95.2% | 102.3% |
| Request 76: GET /vi/XXXXXXXXXXX/oardefault.jpg?... | 1024 | 1723 | 1168 | 1109 | 59.4% | 87.7% | 92.3% |
| Request 77: GET /dist/react-assets/fe21b530ba74... | 2635 | 3100 | 2170 | 2101 | 85.0% | 121.4% | 125.4% |
| Request 78: GET /vi/YYYYYYYYYYY/oardefault.jpg?... | 1019 | 1713 | 1161 | 1102 | 59.5% | 87.8% | 92.5% |
| Request 79: GET /vi/ZZZZZZZZZZZ/oardefault.jpg?... | 1025 | 1724 | 1168 | 1109 | 59.5% | 87.8% | 92.4% |
| Request 80: GET /ogw/aaaaaaaaaaaaaaaaaaaaaaaaaa... | 520 | 955 | 590 | 505 | 54.5% | 88.1% | 103.0% |
| Request 81: GET /images/searchbox/desktop_searc... | 1704 | 2145 | 1446 | 1402 | 79.4% | 117.8% | 121.5% |
| Request 82: GET /u/123456?s=96&v=4 | 839 | 1536 | 991 | 888 | 54.6% | 84.7% | 94.5% |
| Request 83: GET /dist/react-assets/09ab90ee860b... | 2635 | 3100 | 2170 | 2101 | 85.0% | 121.4% | 125.4% |
| Request 84: GET /images?q=aaaaaaaaaaaaaaaaaaaaa... | 830 | 1479 | 968 | 894 | 56.1% | 85.7% | 92.8% |
| Request 85: GET /vi/aaaaaaaaaaaaaaaaaaaaaaaaaaa... | 1014 | 1714 | 1155 | 1094 | 59.2% | 87.8% | 92.7% |
| Request 86: GET /vi/aaaaaaaaaaaaaaaaaaaaaaaaaaa... | 1016 | 1713 | 1156 | 1097 | 59.3% | 87.9% | 92.6% |
| Request 87: GET /images/nav_logo321.webp | 1692 | 2321 | 1562 | 1519 | 72.9% | 108.3% | 111.4% |
| Request 88: GET /images?q=aaaaaaaaaaaaaaaaaaaaa... | 843 | 1491 | 976 | 903 | 56.5% | 86.4% | 93.4% |
| Request 89: GET /ytc/aaaaaaaaaaaaaaaaaaaaaaaaaa... | 776 | 1331 | 884 | 880 | 58.3% | 87.8% | 88.2% |
| Request 90: GET /u/789012?s=40&v=4 | 833 | 1531 | 980 | 876 | 54.4% | 85.0% | 95.1% |
| Request 91: POST /t/ias_videos?aaaaaaaaaaaaaaaa... | 860 | 1044 | 650 | 644 | 82.4% | 132.3% | 133.5% |
| Request 92: GET /generate_204 | 695 | 1113 | 796 | 788 | 62.4% | 87.3% | 88.2% |
| Request 93: OPTIONS /models | 480 | 990 | 628 | 481 | 48.5% | 76.4% | 99.8% |
| Request 94: GET /api/aaaaaaaaaaaaaaaaaaaaaaaaaa... | 1849 | 2226 | 1516 | 1507 | 83.1% | 122.0% | 122.7% |
| Request 95: POST /t/perf?aaaaaaaaaaaaaaaaaaaaaa... | 926 | 1110 | 691 | 685 | 83.4% | 134.0% | 135.2% |
| Request 96: GET /gen_204?s=aaaaaaaaaaaaaaaaaaaa... | 762 | 1008 | 694 | 687 | 75.6% | 109.8% | 110.9% |
| Request 97: POST /t/ias_wikinlp?aaaaaaaaaaaaaaa... | 906 | 1088 | 679 | 673 | 83.3% | 133.4% | 134.6% |
| Request 98: POST /_private/browser/stats | 272 | 494 | 298 | 293 | 55.1% | 91.3% | 92.8% |
| Request 99: POST /t/ias_web?aaaaaaaaaaaaaaaaaaa... | 855 | 1039 | 647 | 641 | 82.3% | 132.1% | 133.4% |
| Request 100: HEAD /generate_204 | 698 | 1104 | 790 | 770 | 63.2% | 88.4% | 90.6% |

</details>

### Request/Response Breakdown

**Request Headers:**

- QH avg: **512 B** (baseline)
- HTTP/1 avg: **775 B** (51.5% larger)
- HTTP/2 avg: **533 B** (4.2% larger)
- HTTP/3 avg: **524 B** (2.4% larger)

**Response Headers:**

- QH avg: **890 B** (baseline)
- HTTP/1 avg: **1167 B** (31.1% larger)
- HTTP/2 avg: **762 B** (14.4% smaller)
- HTTP/3 avg: **714 B** (19.8% smaller)

**Total Headers (Request + Response):**

- QH avg: **1402 B** (baseline)
- HTTP/1 avg: **1942 B** (38.5% larger)
- HTTP/2 avg: **1295 B** (7.6% smaller)
- HTTP/3 avg: **1238 B** (11.7% smaller)

## Combined Results

### Summary

- **110** test cases
- QH total: **144951 B**
- HTTP/1.1 total: **201244 B** (38.8% larger)
- HTTP/2 total: **133914 B** (7.6% smaller)
- HTTP/3 total: **127907 B** (11.8% smaller)

### Performance Bounds

**vs HTTP/1.1:**

- **Best case**: 92.0% smaller - QH: 49 B vs HTTP/1.1: 610 B (Edge Case 1: QH Best Case - All Static Table Complete Pairs)
- **Worst case**: 7.9% smaller - QH: 4196 B vs HTTP/1.1: 4557 B (Request 1: GET /manifest.json)

**vs HTTP/2:**

- **Best case**: 85.5% smaller - QH: 49 B vs HTTP/2: 337 B (Edge Case 1: QH Best Case - All Static Table Complete Pairs)
- **Worst case**: 34.0% larger - QH: 926 B vs HTTP/2: 691 B (Request 95: POST /t/perf?aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa)

**vs HTTP/3:**

- **Best case**: 82.1% smaller - QH: 49 B vs HTTP/3: 274 B (Edge Case 1: QH Best Case - All Static Table Complete Pairs)
- **Worst case**: 36.9% larger - QH: 4196 B vs HTTP/3: 3066 B (Request 1: GET /manifest.json)

### Breakdown by Size Category

| Category | Count | QH Avg | H1 Avg | H2 Avg | H3 Avg | QH/H1 | QH/H2 | QH/H3 |
|----------|------:|-------:|-------:|-------:|-------:|------:|------:|------:|
| Tiny (<1KB) | 75 | 663 B | 1.1 KB | 717 B | 674 B | 58.9% | 92.4% | 98.4% |
| Small (1-10KB) | 35 | 2.7 KB | 3.3 KB | 2.2 KB | 2.2 KB | 81.5% | 118.9% | 123.1% |

### Request/Response Breakdown

**Request Headers:**

- QH avg: **490 B** (baseline)
- HTTP/1 avg: **742 B** (51.3% larger)
- HTTP/2 avg: **508 B** (3.6% larger)
- HTTP/3 avg: **499 B** (1.7% larger)

**Response Headers:**

- QH avg: **827 B** (baseline)
- HTTP/1 avg: **1087 B** (31.4% larger)
- HTTP/2 avg: **709 B** (14.3% smaller)
- HTTP/3 avg: **664 B** (19.7% smaller)

**Total Headers (Request + Response):**

- QH avg: **1318 B** (baseline)
- HTTP/1 avg: **1829 B** (38.8% larger)
- HTTP/2 avg: **1217 B** (7.6% smaller)
- HTTP/3 avg: **1163 B** (11.8% smaller)

## Wire Format Examples

### Example 1: Request 1: GET /manifest.json

**Request Sizes:**

- QH: 280 bytes
- HTTP/1.1: 397 bytes
- HTTP/2: 275 bytes
- HTTP/3: 273 bytes

**QH Request Wire Format (annotated):**

```
OFFSET  BYTES                                            DESCRIPTION
0x0000  00                                               First byte (Version=0, Method=GET)
0x0001  0a                                               Host length: 10
0x0002  67 69 74 68 75 62 2e 63 6f 6d                    Host: github.com
0x000c  0e                                               Path length: 14
0x000d  2f 6d 61 6e 69 66 65 73 74 2e 6a 73 6f 6e        Path: /manifest.json
0x001b  fa 01                                            Headers length: 250

0x001d  41                                               Header ID (user-agent)
0x001e  75                                                 Value length: 117
0x001f  4d 6f 7a 69 6c 6c 61 2f 35 2e 30 20 28 4d 61 63    Value: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.... [117 bytes total, showing 100]
          69 6e 74 6f 73 68 3b 20 49 6e 74 65 6c 20 4d 61
          63 20 4f 53 20 58 20 31 30 5f 31 35 5f 37 29 20
          41 70 70 6c 65 57 65 62 4b 69 74 2f 35 33 37 2e
          33 36 20 28 4b 48 54 4d 4c 2c 20 6c 69 6b 65 20
          47 65 63 6b 6f 29 20 43 68 72 6f 6d 65 2f 31 34
          32 2e 30 2e 30 2e 30 20 53 61 66 61 72 69 2f 35
          33 37 2e 33 36
0x0094  01                                               Header ID (sec-ch-ua-mobile: ?0)
0x0095  3d                                               Header ID (sec-ch-ua-platform: "macOS")
0x0096  51                                               Header ID (referer)
0x0097  2c                                                 Value length: 44
0x0098  68 74 74 70 73 3a 2f 2f 67 69 74 68 75 62 2e 63    Value: https://github.com/example-user/example-repo
          6f 6d 2f 65 78 61 6d 70 6c 65 2d 75 73 65 72 2f
          65 78 61 6d 70 6c 65 2d 72 65 70 6f
0x00c4  4c                                               Header ID (accept-language)
0x00c5  0e                                                 Value length: 14
0x00c6  65 6e 2d 47 42 2c 65 6e 3b 71 3d 30 2e 39          Value: en-GB,en;q=0.9
0x00d4  44                                               Header ID (sec-ch-ua)
0x00d5  41                                                 Value length: 65
0x00d6  22 43 68 72 6f 6d 69 75 6d 22 3b 76 3d 22 31 34    Value: "Chromium";v="142", "Google Chrome";v="142", "Not_A Brand";v="99"
          32 22 2c 20 22 47 6f 6f 67 6c 65 20 43 68 72 6f
          6d 65 22 3b 76 3d 22 31 34 32 22 2c 20 22 4e 6f
          74 5f 41 20 42 72 61 6e 64 22 3b 76 3d 22 39 39
          22

0x0117  00                                               Body length: 0

Summary: parsed 280 / 280 bytes
```

**Response Sizes:**

- QH: 3916 bytes
- HTTP/1.1: 4160 bytes
- HTTP/2: 2868 bytes
- HTTP/3: 2793 bytes

**QH Response Wire Format (annotated):**

```
OFFSET  BYTES                                            DESCRIPTION
0x0000  00                                               First byte (Version=0, Status=200)
0x0001  c8 1e                                            Headers length: 3912

0x0003  53                                               Header ID (x-frame-options: deny)
0x0004  05                                               Header ID (content-encoding: gzip)
0x0005  00                                               Custom header
0x0006  13                                                 Key length: 19
0x0007  78 2d 67 69 74 68 75 62 2d 72 65 71 75 65 73 74    Key: x-github-request-id
          2d 69 64
0x001a  1f                                                 Value length: 31
0x001b  58 58 58 58 3a 58 58 58 58 3a 58 58 58 58 58 58    Value: XXXX:XXXX:XXXXXX:XXXXX:XXXXXXXX
          3a 58 58 58 58 58 3a 58 58 58 58 58 58 58 58
0x003a  8e                                               Header ID (date)
0x003b  1d                                                 Value length: 29
0x003c  53 61 74 2c 20 31 35 20 4e 6f 76 20 32 30 32 35    Value: Sat, 15 Nov 2025 17:15:49 GMT
          20 31 37 3a 31 35 3a 34 39 20 47 4d 54
0x0059  0e                                               Header ID (x-xss-protection: 0)
0x005a  ac                                               Header ID (referrer-policy)
0x005b  39                                                 Value length: 57
0x005c  6f 72 69 67 69 6e 2d 77 68 65 6e 2d 63 72 6f 73    Value: origin-when-cross-origin, strict-origin-when-cross-origin
          73 2d 6f 72 69 67 69 6e 2c 20 73 74 72 69 63 74
          2d 6f 72 69 67 69 6e 2d 77 68 65 6e 2d 63 72 6f
          73 73 2d 6f 72 69 67 69 6e
0x0095  a1                                               Header ID (etag)
0x0096  26                                                 Value length: 38
0x0097  61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61    Value: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
          61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61
          61 61 61 61 61 61
0x00bd  01                                               Header ID (x-content-type-options: nosniff)
0x00be  0b                                               Header ID (accept-ranges: bytes)
0x00bf  91                                               Header ID (content-length)
0x00c0  03                                                 Value length: 3
0x00c1  34 37 34                                           Value: 474
0x00c4  ab                                               Header ID (content-security-policy)
0x00c5  b7 1b                                              Value length: 3511
0x00c7  64 65 66 61 75 6c 74 2d 73 72 63 20 27 6e 6f 6e    Value: default-src 'none'; base-uri 'self'; child-src github.githubassets.com github.com/assets-cdn/worker/... [3511 bytes total, showing 100]
          65 27 3b 20 62 61 73 65 2d 75 72 69 20 27 73 65
          6c 66 27 3b 20 63 68 69 6c 64 2d 73 72 63 20 67
          69 74 68 75 62 2e 67 69 74 68 75 62 61 73 73 65
          74 73 2e 63 6f 6d 20 67 69 74 68 75 62 2e 63 6f
          6d 2f 61 73 73 65 74 73 2d 63 64 6e 2f 77 6f 72
          6b 65 72 2f 20 67 69 74 68 75 62 2e 63 6f 6d 2f
          61 73 73 65 74 73 2f 20 67 69 73 74 2e 67 69 74
                                                         [3511 bytes total; 128 shown]
0x0e7e  90                                               Header ID (cache-control)
0x0e7f  16                                                 Value length: 22
0x0e80  6d 61 78 2d 61 67 65 3d 36 30 34 38 30 30 2c 20    Value: max-age=604800, public
          70 75 62 6c 69 63
0x0e96  8f                                               Header ID (content-type)
0x0e97  28                                                 Value length: 40
0x0e98  61 70 70 6c 69 63 61 74 69 6f 6e 2f 6d 61 6e 69    Value: application/manifest+json; charset=utf-8
          66 65 73 74 2b 6a 73 6f 6e 3b 20 63 68 61 72 73
          65 74 3d 75 74 66 2d 38
0x0ec0  95                                               Header ID (vary)
0x0ec1  7d                                                 Value length: 125
0x0ec2  58 2d 46 65 74 63 68 2d 4e 6f 6e 63 65 2c 20 58    Value: X-Fetch-Nonce, X-PJAX, X-PJAX-Container, Turbo-Visit, Turbo-Frame, X-Requested-With,Accept-Encoding,... [125 bytes total, showing 100]
          2d 50 4a 41 58 2c 20 58 2d 50 4a 41 58 2d 43 6f
          6e 74 61 69 6e 65 72 2c 20 54 75 72 62 6f 2d 56
          69 73 69 74 2c 20 54 75 72 62 6f 2d 46 72 61 6d
          65 2c 20 58 2d 52 65 71 75 65 73 74 65 64 2d 57
          69 74 68 2c 41 63 63 65 70 74 2d 45 6e 63 6f 64
          69 6e 67 2c 20 41 63 63 65 70 74 2c 20 58 2d 52
          65 71 75 65 73 74 65 64 2d 57 69 74 68
0x0f3f  92                                               Header ID (server)
0x0f40  0a                                                 Value length: 10
0x0f41  67 69 74 68 75 62 2e 63 6f 6d                      Value: github.com

0x0f4b  00                                               Body length: 0

Summary: parsed 3916 / 3916 bytes
```

### Example 2: Request 2: POST /github-copilot/chat/token

**Request Sizes:**

- QH: 443 bytes
- HTTP/1.1: 792 bytes
- HTTP/2: 534 bytes
- HTTP/3: 514 bytes

**QH Request Wire Format (annotated):**

```
OFFSET  BYTES                                            DESCRIPTION
0x0000  08                                               First byte (Version=0, Method=POST)
0x0001  0a                                               Host length: 10
0x0002  67 69 74 68 75 62 2e 63 6f 6d                    Host: github.com
0x000c  1a                                               Path length: 26
0x000d  2f 67 69 74 68 75 62 2d 63 6f 70 69 6c 6f 74 2f  Path: /github-copilot/chat/token
          63 68 61 74 2f 74 6f 6b 65 6e
0x0027  91 03                                            Headers length: 401

0x0029  00                                               Custom header
0x002a  15                                                 Key length: 21
0x002b  67 69 74 68 75 62 2d 76 65 72 69 66 69 65 64 2d    Key: github-verified-fetch
          66 65 74 63 68
0x0040  04                                                 Value length: 4
0x0041  74 72 75 65                                        Value: true
0x0045  16                                               Header ID (sec-fetch-dest: empty)
0x0046  00                                               Custom header
0x0047  17                                                 Key length: 23
0x0048  78 2d 67 69 74 68 75 62 2d 63 6c 69 65 6e 74 2d    Key: x-github-client-version
          76 65 72 73 69 6f 6e
0x005f  1b                                                 Value length: 27
0x0060  41 4e 4f 4e 59 4d 49 5a 45 44 5f 43 4c 49 45 4e    Value: ANONYMIZED_CLIENT_VERSION_1
          54 5f 56 45 52 53 49 4f 4e 5f 31
0x007b  4c                                               Header ID (accept-language)
0x007c  0e                                                 Value length: 14
0x007d  65 6e 2d 47 42 2c 65 6e 3b 71 3d 30 2e 39          Value: en-GB,en;q=0.9
0x008b  53                                               Header ID (origin)
0x008c  12                                                 Value length: 18
0x008d  68 74 74 70 73 3a 2f 2f 67 69 74 68 75 62 2e 63    Value: https://github.com
          6f 6d
0x009f  51                                               Header ID (referer)
0x00a0  1a                                                 Value length: 26
0x00a1  68 74 74 70 73 3a 2f 2f 67 69 74 68 75 62 2e 63    Value: https://github.com/example
          6f 6d 2f 65 78 61 6d 70 6c 65
0x00bb  01                                               Header ID (sec-ch-ua-mobile: ?0)
0x00bc  41                                               Header ID (user-agent)
0x00bd  75                                                 Value length: 117
0x00be  4d 6f 7a 69 6c 6c 61 2f 35 2e 30 20 28 4d 61 63    Value: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.... [117 bytes total, showing 100]
          69 6e 74 6f 73 68 3b 20 49 6e 74 65 6c 20 4d 61
          63 20 4f 53 20 58 20 31 30 5f 31 35 5f 37 29 20
          41 70 70 6c 65 57 65 62 4b 69 74 2f 35 33 37 2e
          33 36 20 28 4b 48 54 4d 4c 2c 20 6c 69 6b 65 20
          47 65 63 6b 6f 29 20 43 68 72 6f 6d 65 2f 31 34
          32 2e 30 2e 30 2e 30 20 53 61 66 61 72 69 2f 35
          33 37 2e 33 36
0x0133  00                                               Custom header
0x0134  0d                                                 Key length: 13
0x0135  78 2d 66 65 74 63 68 2d 6e 6f 6e 63 65             Key: x-fetch-nonce
0x0142  1b                                                 Value length: 27
0x0143  76 32 3a 41 4e 4f 4e 59 4d 49 5a 45 44 5f 4e 4f    Value: v2:ANONYMIZED_NONCE_VALUE_1
          4e 43 45 5f 56 41 4c 55 45 5f 31
0x015e  0d                                               Header ID (accept-encoding: gzip, deflate, br, zstd)
0x015f  0e                                               Header ID (content-type: application/json)
0x0160  00                                               Custom header
0x0161  08                                                 Key length: 8
0x0162  70 72 69 6f 72 69 74 79                            Key: priority
0x016a  06                                                 Value length: 6
0x016b  75 3d 31 2c 20 69                                  Value: u=1, i
0x0171  44                                               Header ID (sec-ch-ua)
0x0172  41                                                 Value length: 65
0x0173  22 43 68 72 6f 6d 69 75 6d 22 3b 76 3d 22 31 34    Value: "Chromium";v="142", "Google Chrome";v="142", "Not_A Brand";v="99"
          32 22 2c 20 22 47 6f 6f 67 6c 65 20 43 68 72 6f
          6d 65 22 3b 76 3d 22 31 34 32 22 2c 20 22 4e 6f
          74 5f 41 20 42 72 61 6e 64 22 3b 76 3d 22 39 39
          22
0x01b4  1d                                               Header ID (accept: application/json)
0x01b5  33                                               Header ID (content-length: 0)
0x01b6  3d                                               Header ID (sec-ch-ua-platform: "macOS")
0x01b7  0f                                               Header ID (sec-fetch-mode: cors)
0x01b8  12                                               Header ID (sec-fetch-site: same-origin)
0x01b9  05                                               Header ID (x-requested-with: XMLHttpRequest)

0x01ba  00                                               Body length: 0

Summary: parsed 443 / 443 bytes
```

**Response Sizes:**

- QH: 3892 bytes
- HTTP/1.1: 4215 bytes
- HTTP/2: 2894 bytes
- HTTP/3: 2825 bytes

**QH Response Wire Format (annotated):**

```
OFFSET  BYTES                                            DESCRIPTION
0x0000  00                                               First byte (Version=0, Status=200)
0x0001  b0 1e                                            Headers length: 3888

0x0003  94                                               Header ID (strict-transport-security)
0x0004  2c                                                 Value length: 44
0x0005  6d 61 78 2d 61 67 65 3d 33 31 35 33 36 30 30 30    Value: max-age=31536000; includeSubdomains; preload
          3b 20 69 6e 63 6c 75 64 65 53 75 62 64 6f 6d 61
          69 6e 73 3b 20 70 72 65 6c 6f 61 64
0x0031  8e                                               Header ID (date)
0x0032  1d                                                 Value length: 29
0x0033  53 61 74 2c 20 31 35 20 4e 6f 76 20 32 30 32 35    Value: Sat, 15 Nov 2025 17:15:50 GMT
          20 31 37 3a 31 35 3a 35 30 20 47 4d 54
0x0050  95                                               Header ID (vary)
0x0051  7d                                                 Value length: 125
0x0052  58 2d 46 65 74 63 68 2d 4e 6f 6e 63 65 2c 20 58    Value: X-Fetch-Nonce, X-PJAX, X-PJAX-Container, Turbo-Visit, Turbo-Frame, X-Requested-With,Accept-Encoding,... [125 bytes total, showing 100]
          2d 50 4a 41 58 2c 20 58 2d 50 4a 41 58 2d 43 6f
          6e 74 61 69 6e 65 72 2c 20 54 75 72 62 6f 2d 56
          69 73 69 74 2c 20 54 75 72 62 6f 2d 46 72 61 6d
          65 2c 20 58 2d 52 65 71 75 65 73 74 65 64 2d 57
          69 74 68 2c 41 63 63 65 70 74 2d 45 6e 63 6f 64
          69 6e 67 2c 20 41 63 63 65 70 74 2c 20 58 2d 52
          65 71 75 65 73 74 65 64 2d 57 69 74 68
0x00cf  01                                               Header ID (x-content-type-options: nosniff)
0x00d0  53                                               Header ID (x-frame-options: deny)
0x00d1  00                                               Custom header
0x00d2  13                                                 Key length: 19
0x00d3  78 2d 67 69 74 68 75 62 2d 72 65 71 75 65 73 74    Key: x-github-request-id
          2d 69 64
0x00e6  1f                                                 Value length: 31
0x00e7  58 58 58 58 3a 58 58 58 58 3a 58 58 58 58 58 58    Value: XXXX:XXXX:XXXXXX:XXXXX:XXXXXXXX
          3a 58 58 58 58 58 3a 58 58 58 58 58 58 58 58
0x0106  0e                                               Header ID (x-xss-protection: 0)
0x0107  3e                                               Header ID (cache-control: max-age=0, private, must-revalidate)
0x0108  05                                               Header ID (content-encoding: gzip)
0x0109  ab                                               Header ID (content-security-policy)
0x010a  b7 1b                                              Value length: 3511
0x010c  64 65 66 61 75 6c 74 2d 73 72 63 20 27 6e 6f 6e    Value: default-src 'none'; base-uri 'self'; child-src github.githubassets.com github.com/assets-cdn/worker/... [3511 bytes total, showing 100]
          65 27 3b 20 62 61 73 65 2d 75 72 69 20 27 73 65
          6c 66 27 3b 20 63 68 69 6c 64 2d 73 72 63 20 67
          69 74 68 75 62 2e 67 69 74 68 75 62 61 73 73 65
          74 73 2e 63 6f 6d 20 67 69 74 68 75 62 2e 63 6f
          6d 2f 61 73 73 65 74 73 2d 63 64 6e 2f 77 6f 72
          6b 65 72 2f 20 67 69 74 68 75 62 2e 63 6f 6d 2f
          61 73 73 65 74 73 2f 20 67 69 73 74 2e 67 69 74
                                                         [3511 bytes total; 128 shown]
0x0ec3  0c                                               Header ID (content-type: application/json; charset=utf-8)
0x0ec4  a1                                               Header ID (etag)
0x0ec5  26                                                 Value length: 38
0x0ec6  61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61    Value: aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
          61 61 61 61 61 61 61 61 61 61 61 61 61 61 61 61
          61 61 61 61 61 61
0x0eec  ac                                               Header ID (referrer-policy)
0x0eed  39                                                 Value length: 57
0x0eee  6f 72 69 67 69 6e 2d 77 68 65 6e 2d 63 72 6f 73    Value: origin-when-cross-origin, strict-origin-when-cross-origin
          73 2d 6f 72 69 67 69 6e 2c 20 73 74 72 69 63 74
          2d 6f 72 69 67 69 6e 2d 77 68 65 6e 2d 63 72 6f
          73 73 2d 6f 72 69 67 69 6e
0x0f27  92                                               Header ID (server)
0x0f28  0a                                                 Value length: 10
0x0f29  67 69 74 68 75 62 2e 63 6f 6d                      Value: github.com

0x0f33  00                                               Body length: 0

Summary: parsed 3892 / 3892 bytes
```

### Example 3: Request 3: GET /_next/data/qcPKpvHcKq6nT9zVikeiO/docs.json

**Request Sizes:**

- QH: 418 bytes
- HTTP/1.1: 696 bytes
- HTTP/2: 460 bytes
- HTTP/3: 442 bytes

**QH Request Wire Format (annotated):**

```
OFFSET  BYTES                                            DESCRIPTION
0x0000  00                                               First byte (Version=0, Method=GET)
0x0001  0b                                               Host length: 11
0x0002  67 68 6f 73 74 74 79 2e 6f 72 67                 Host: ghostty.org
0x000d  29                                               Path length: 41
0x000e  2f 5f 6e 65 78 74 2f 64 61 74 61 2f 41 4e 4f 4e  Path: /_next/data/ANONYMIZED_BUILD_ID/docs.json
          59 4d 49 5a 45 44 5f 42 55 49 4c 44 5f 49 44 2f
          64 6f 63 73 2e 6a 73 6f 6e
0x0037  e8 02                                            Headers length: 360

0x0039  51                                               Header ID (referer)
0x003a  1e                                                 Value length: 30
0x003b  68 74 74 70 73 3a 2f 2f 67 68 6f 73 74 74 79 2e    Value: https://ghostty.org/docs/about
          6f 72 67 2f 64 6f 63 73 2f 61 62 6f 75 74
0x0059  44                                               Header ID (sec-ch-ua)
0x005a  41                                                 Value length: 65
0x005b  22 43 68 72 6f 6d 69 75 6d 22 3b 76 3d 22 31 34    Value: "Chromium";v="142", "Google Chrome";v="142", "Not_A Brand";v="99"
          32 22 2c 20 22 47 6f 6f 67 6c 65 20 43 68 72 6f
          6d 65 22 3b 76 3d 22 31 34 32 22 2c 20 22 4e 6f
          74 5f 41 20 42 72 61 6e 64 22 3b 76 3d 22 39 39
          22
0x009c  16                                               Header ID (sec-fetch-dest: empty)
0x009d  12                                               Header ID (sec-fetch-site: same-origin)
0x009e  63                                               Header ID (if-modified-since)
0x009f  1d                                                 Value length: 29
0x00a0  54 68 75 2c 20 32 33 20 4f 63 74 20 32 30 32 35    Value: Thu, 23 Oct 2025 17:58:54 GMT
          20 31 37 3a 35 38 3a 35 34 20 47 4d 54
0x00bd  00                                               Custom header
0x00be  08                                                 Key length: 8
0x00bf  70 72 69 6f 72 69 74 79                            Key: priority
0x00c7  06                                                 Value length: 6
0x00c8  75 3d 31 2c 20 69                                  Value: u=1, i
0x00ce  00                                               Custom header
0x00cf  07                                                 Key length: 7
0x00d0  70 75 72 70 6f 73 65                               Key: purpose
0x00d7  08                                                 Value length: 8
0x00d8  70 72 65 66 65 74 63 68                            Value: prefetch
0x00e0  03                                               Header ID (accept: */*)
0x00e1  60                                               Header ID (if-none-match)
0x00e2  22                                                 Value length: 34
0x00e3  22 36 30 62 31 39 31 32 33 39 38 37 31 32 33 61    Value: "60b19123987123asdlk3a57794917152"
          73 64 6c 6b 33 61 35 37 37 39 34 39 31 37 31 35
          32 22
0x0105  01                                               Header ID (sec-ch-ua-mobile: ?0)
0x0106  3d                                               Header ID (sec-ch-ua-platform: "macOS")
0x0107  41                                               Header ID (user-agent)
0x0108  75                                                 Value length: 117
0x0109  4d 6f 7a 69 6c 6c 61 2f 35 2e 30 20 28 4d 61 63    Value: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.... [117 bytes total, showing 100]
          69 6e 74 6f 73 68 3b 20 49 6e 74 65 6c 20 4d 61
          63 20 4f 53 20 58 20 31 30 5f 31 35 5f 37 29 20
          41 70 70 6c 65 57 65 62 4b 69 74 2f 35 33 37 2e
          33 36 20 28 4b 48 54 4d 4c 2c 20 6c 69 6b 65 20
          47 65 63 6b 6f 29 20 43 68 72 6f 6d 65 2f 31 34
          32 2e 30 2e 30 2e 30 20 53 61 66 61 72 69 2f 35
          33 37 2e 33 36
0x017e  00                                               Custom header
0x017f  0d                                                 Key length: 13
0x0180  78 2d 6e 65 78 74 6a 73 2d 64 61 74 61             Key: x-nextjs-data
0x018d  01                                                 Value length: 1
0x018e  31                                                 Value: 1
0x018f  4c                                               Header ID (accept-language)
0x0190  0e                                                 Value length: 14
0x0191  65 6e 2d 47 42 2c 65 6e 3b 71 3d 30 2e 39          Value: en-GB,en;q=0.9
0x019f  0f                                               Header ID (sec-fetch-mode: cors)
0x01a0  0d                                               Header ID (accept-encoding: gzip, deflate, br, zstd)

0x01a1  00                                               Body length: 0

Summary: parsed 418 / 418 bytes
```

**Response Sizes:**

- QH: 85 bytes
- HTTP/1.1: 176 bytes
- HTTP/2: 103 bytes
- HTTP/3: 101 bytes

**QH Response Wire Format (annotated):**

```
OFFSET  BYTES                                            DESCRIPTION
0x0000  08                                               First byte (Version=0, Status=304)
0x0001  52                                               Headers length: 82

0x0002  79                                               Header ID (cache-control: public, max-age=0, must-revalidate)
0x0003  8e                                               Header ID (date)
0x0004  1d                                                 Value length: 29
0x0005  53 61 74 2c 20 31 35 20 4e 6f 76 20 32 30 32 35    Value: Sat, 15 Nov 2025 17:16:16 GMT
          20 31 37 3a 31 36 3a 31 36 20 47 4d 54
0x0022  92                                               Header ID (server)
0x0023  06                                                 Value length: 6
0x0024  56 65 72 63 65 6c                                  Value: Vercel
0x002a  00                                               Custom header
0x002b  0b                                                 Key length: 11
0x002c  78 2d 76 65 72 63 65 6c 2d 69 64                   Key: x-vercel-id
0x0037  1c                                                 Value length: 28
0x0038  66 72 61 31 3a 3a 41 4e 4f 4e 59 4d 49 5a 45 44    Value: fra1::ANONYMIZED_VERCEL_ID_1
          5f 56 45 52 43 45 4c 5f 49 44 5f 31

0x0054  00                                               Body length: 0

Summary: parsed 85 / 85 bytes
```

