import fnmatch
import os
from functools import partial
from pathlib import Path
import json
from typing import Optional
from ctypes import c_uint64
import pytest
from eth_account.messages import encode_typed_data
import web3

from ledgered.devices import Device

from ragger.backend import BackendInterface
from ragger.firmware.touch.positions import POSITIONS
from ragger.navigator import Navigator, NavInsID, NavIns
from ragger.error import ExceptionRAPDU

import client.response_parser as ResponseParser
from client.utils import recover_message
from client.client import EthAppClient, TrustedNameType, TrustedNameSource
from client.status_word import StatusWord
from client.eip712 import InputData
from client.settings import SettingID, settings_toggle
from client.tx_simu import TxSimu
from client.proxy_info import ProxyInfo


BIP32_PATH = "m/44'/60'/0'/0/0"
autonext_idx: int
snapshots_dirname: Optional[str] = None
WALLET_ADDR: Optional[bytes] = None
validate_warning: bool = False
skip_flow: bool = False


def eip712_json_path() -> str:
    return f"{os.path.dirname(__file__)}/eip712_input_files"


def input_files() -> list[str]:
    files = []
    for file in os.scandir(eip712_json_path()):
        if fnmatch.fnmatch(file, "*-data.json"):
            files.append(file.path)
    return sorted(files)


@pytest.fixture(name="input_file", params=input_files())
def input_file_fixture(request) -> Path:
    return Path(request.param)


@pytest.fixture(name="verbose_raw", params=[True, False])
def verbose_raw_fixture(request) -> bool:
    return request.param


@pytest.fixture(name="filtering", params=[False, True])
def filtering_fixture(request) -> bool:
    return request.param


def get_wallet_addr(client: EthAppClient) -> bytes:
    global WALLET_ADDR

    # don't ask again if we already have it
    if WALLET_ADDR is None:
        with client.get_public_addr(display=False):
            pass
        _, WALLET_ADDR, _ = ResponseParser.pk_addr(client.response().data)
    return WALLET_ADDR


def test_eip712_v0(backend: BackendInterface,
                   navigator: Navigator,
                   simu_params: Optional[TxSimu] = None):
    global validate_warning

    app_client = EthAppClient(backend)
    device = backend.device

    DEVICE_ADDR = get_wallet_addr(app_client)

    settings_toggle(device, navigator, [SettingID.BLIND_SIGNING])
    with open(input_files()[0], encoding="utf-8") as file:
        data = json.load(file)
    smsg = encode_typed_data(full_message=data)

    if simu_params is not None:
        validate_warning = True
        simu_params.from_addr = DEVICE_ADDR
        simu_params.tx_hash = smsg.body
        simu_params.domain_hash = smsg.header
        response = app_client.provide_tx_simulation(simu_params)
        assert response.status == StatusWord.OK

    with app_client.eip712_sign_legacy(BIP32_PATH, smsg.header, smsg.body):
        moves = []
        if device.is_nano:
            moves += [NavInsID.BOTH_CLICK]
            moves += [NavInsID.RIGHT_CLICK] * 5
            moves += [NavInsID.BOTH_CLICK]
        else:
            moves += [NavInsID.USE_CASE_CHOICE_REJECT]
            moves += [NavInsID.SWIPE_CENTER_TO_LEFT] * 2
            moves += [NavInsID.USE_CASE_REVIEW_CONFIRM]
        navigator.navigate(moves)

    vrs = ResponseParser.signature(app_client.response().data)
    assert DEVICE_ADDR == recover_message(data, vrs)


def autonext(device: Device, navigator: Navigator, default_screenshot_path: Path):
    global autonext_idx

    moves = []
    if device.is_nano:
        if autonext_idx == 0 and validate_warning:
            moves = [NavInsID.BOTH_CLICK]
        else:
            moves = [NavInsID.RIGHT_CLICK]
    else:
        if autonext_idx == 0 and validate_warning:
            moves = [NavInsID.USE_CASE_CHOICE_REJECT]
        else:
            if autonext_idx == 2 and skip_flow:
                InputData.disable_autonext()  # so the timer stops firing
                moves = [
                    # Ragger does not handle the skip button
                    NavIns(NavInsID.TOUCH, POSITIONS["RightHeader"][device.type]),
                    NavInsID.USE_CASE_CHOICE_CONFIRM,
                ]
            else:
                moves = [NavInsID.SWIPE_CENTER_TO_LEFT]
    if snapshots_dirname is not None:
        navigator.navigate_and_compare(default_screenshot_path,
                                       snapshots_dirname,
                                       moves,
                                       screen_change_before_first_instruction=False,
                                       screen_change_after_last_instruction=False,
                                       snap_start_idx=autonext_idx)
    else:
        navigator.navigate(moves,
                           screen_change_before_first_instruction=False,
                           screen_change_after_last_instruction=False)
    autonext_idx += len(moves)


def eip712_new_common(device: Device,
                      navigator: Navigator,
                      default_screenshot_path: Path,
                      app_client: EthAppClient,
                      json_data: dict,
                      filters: Optional[dict],
                      golden_run: bool):
    global autonext_idx
    global validate_warning
    global skip_flow
    global snapshots_dirname

    autonext_idx = 0
    assert InputData.process_data(app_client,
                                  json_data,
                                  filters,
                                  partial(autonext, device, navigator, default_screenshot_path),
                                  golden_run)
    with app_client.eip712_sign_new(BIP32_PATH):
        if device.is_nano:
            nav_ins = NavInsID.RIGHT_CLICK
            val_ins = NavInsID.BOTH_CLICK
            text = "Sign message"
        else:
            nav_ins = NavInsID.SWIPE_CENTER_TO_LEFT
            val_ins = NavInsID.USE_CASE_REVIEW_CONFIRM
            text = "Hold to sign"
        if snapshots_dirname is not None:
            navigator.navigate_until_text_and_compare(nav_ins,
                                                      [val_ins],
                                                      text,
                                                      default_screenshot_path,
                                                      snapshots_dirname,
                                                      snap_start_idx=autonext_idx)
        else:
            navigator.navigate_until_text(nav_ins, [val_ins], text)
    # reset values
    validate_warning = False
    skip_flow = False
    snapshots_dirname = None

    return ResponseParser.signature(app_client.response().data)


def get_filter_file_from_data_file(data_file: Path) -> Path:
    test_path = f"{data_file.parent}/{'-'.join(data_file.stem.split('-')[:-1])}"
    return Path(f"{test_path}-filter.json")


def test_eip712_new(backend: BackendInterface,
                    navigator: Navigator,
                    default_screenshot_path: Path,
                    input_file: Path,
                    verbose_raw: bool,
                    filtering: bool):
    global validate_warning

    settings_to_toggle: list[SettingID] = []
    app_client = EthAppClient(backend)
    device = backend.device

    filters = None
    if filtering:
        try:
            filterfile = get_filter_file_from_data_file(input_file)
            with open(filterfile, encoding="utf-8") as f:
                filters = json.load(f)
        except (IOError, json.decoder.JSONDecodeError) as e:
            pytest.skip(f"{filterfile.name}: {e.strerror}")
    else:
        settings_to_toggle.append(SettingID.BLIND_SIGNING)

    if verbose_raw:
        settings_to_toggle.append(SettingID.VERBOSE_EIP712)

    if not filters or verbose_raw:
        validate_warning = True

    if len(settings_to_toggle) > 0:
        settings_toggle(device, navigator, settings_to_toggle)

    with open(input_file, encoding="utf-8") as file:
        data = json.load(file)
        vrs = eip712_new_common(device,
                                navigator,
                                default_screenshot_path,
                                app_client,
                                data,
                                filters,
                                False)

        recovered_addr = recover_message(data, vrs)

    assert recovered_addr == get_wallet_addr(app_client)


class DataSet():
    data: dict
    filters: dict
    suffix: str

    def __init__(self, data: dict, filters: dict, suffix: str = ""):
        self.data = data
        self.filters = filters
        self.suffix = suffix


ADVANCED_DATA_SETS = [
    DataSet(
        {
            "domain": {
                "chainId": 1,
                "name": "Advanced test",
                "verifyingContract": "0xCcCCccccCCCCcCCCCCCcCcCccCcCCCcCcccccccC",
                "version": "1"
            },
            "message": {
                "with": "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045",
                "value_recv": 10000000000000000,
                "token_send": "0x6B175474E89094C44Da98b954EedeAC495271d0F",
                "value_send": 24500000000000000000,
                "token_recv": "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
                "expires": 1714559400,
            },
            "primaryType": "Transfer",
            "types": {
                "EIP712Domain": [
                    {"name": "name", "type": "string"},
                    {"name": "version", "type": "string"},
                    {"name": "chainId", "type": "uint256"},
                    {"name": "verifyingContract", "type": "address"}
                ],
                "Transfer": [
                    {"name": "with", "type": "address"},
                    {"name": "value_recv", "type": "uint256"},
                    {"name": "token_send", "type": "address"},
                    {"name": "value_send", "type": "uint256"},
                    {"name": "token_recv", "type": "address"},
                    {"name": "expires", "type": "uint64"},
                ]
            }
        },
        {
            "name": "Advanced Filtering",
            "tokens": [
                {
                    "addr": "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
                    "ticker": "WETH",
                    "decimals": 18,
                    "chain_id": 1,
                },
                {
                    "addr": "0x6b175474e89094c44da98b954eedeac495271d0f",
                    "ticker": "DAI",
                    "decimals": 18,
                    "chain_id": 1,
                },
            ],
            "fields": {
                "value_send": {
                    "type": "amount_join_value",
                    "name": "Send",
                    "token": 1,
                },
                "token_send": {
                    "type": "amount_join_token",
                    "token": 1,
                },
                "value_recv": {
                    "type": "amount_join_value",
                    "name": "Receive",
                    "token": 0,
                },
                "token_recv": {
                    "type": "amount_join_token",
                    "token": 0,
                },
                "with": {
                    "type": "raw",
                    "name": "With",
                },
                "expires": {
                    "type": "datetime",
                    "name": "Will Expire"
                },
            }
        }
    ),
    DataSet(
        {
            "types": {
                "EIP712Domain": [
                    {"name": "name", "type": "string"},
                    {"name": "version", "type": "string"},
                    {"name": "chainId", "type": "uint256"},
                    {"name": "verifyingContract", "type": "address"},
                ],
                "Permit": [
                    {"name": "owner", "type": "address"},
                    {"name": "spender", "type": "address"},
                    {"name": "value", "type": "uint256"},
                    {"name": "nonce", "type": "uint256"},
                    {"name": "deadline", "type": "uint256"},
                ]
            },
            "primaryType": "Permit",
            "domain": {
                "name": "ENS",
                "version": "1",
                "verifyingContract": "0xC18360217D8F7Ab5e7c516566761Ea12Ce7F9D72",
                "chainId": 1,
            },
            "message": {
                "owner": "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045",
                "spender": "0x5B38Da6a701c568545dCfcB03FcB875f56beddC4",
                "value": 4200000000000000000,
                "nonce": 0,
                "deadline": 1719756000,
            }
        },
        {
            "name": "Permit filtering",
            "tokens": [
                {
                    "addr": "0xC18360217D8F7Ab5e7c516566761Ea12Ce7F9D72",
                    "ticker": "ENS",
                    "decimals": 18,
                    "chain_id": 1,
                },
            ],
            "fields": {
                "value": {
                    "type": "amount_join_value",
                    "name": "Send",
                },
                "deadline": {
                    "type": "datetime",
                    "name": "Deadline",
                },
            }
        },
        "_permit"
    ),
    DataSet(
        {
            "types": {
                "EIP712Domain": [
                    {"name": "name", "type": "string"},
                    {"name": "version", "type": "string"},
                    {"name": "chainId", "type": "uint256"},
                    {"name": "verifyingContract", "type": "address"},
                ],
                "Root": [
                    {"name": "token_big", "type": "address"},
                    {"name": "value_big", "type": "uint256"},
                    {"name": "token_biggest", "type": "address"},
                    {"name": "value_biggest", "type": "uint256"},
                ]
            },
            "primaryType": "Root",
            "domain": {
                "name": "test",
                "version": "1",
                "verifyingContract": "0x0000000000000000000000000000000000000000",
                "chainId": 1,
            },
            "message": {
                "token_big": "0x6b175474e89094c44da98b954eedeac495271d0f",
                "value_big": c_uint64(-1).value,
                "token_biggest": "0x6b175474e89094c44da98b954eedeac495271d0f",
                "value_biggest": int(web3.constants.MAX_INT, 0),
            }
        },
        {
            "name": "Unlimited test",
            "tokens": [
                {
                    "addr": "0x6b175474e89094c44da98b954eedeac495271d0f",
                    "ticker": "DAI",
                    "decimals": 18,
                    "chain_id": 1,
                },
            ],
            "fields": {
                "token_big": {
                    "type": "amount_join_token",
                    "token": 0,
                },
                "value_big": {
                    "type": "amount_join_value",
                    "name": "Big",
                    "token": 0,
                },
                "token_biggest": {
                    "type": "amount_join_token",
                    "token": 0,
                },
                "value_biggest": {
                    "type": "amount_join_value",
                    "name": "Biggest",
                    "token": 0,
                },
            }
        },
        "_unlimited"
    ),
]


@pytest.fixture(name="data_set", params=ADVANCED_DATA_SETS)
def data_set_fixture(request) -> DataSet:
    return request.param


def test_eip712_advanced_filtering(backend: BackendInterface,
                                   navigator: Navigator,
                                   default_screenshot_path: Path,
                                   test_name: str,
                                   data_set: DataSet,
                                   golden_run: bool,
                                   verbose_raw: bool):
    global snapshots_dirname

    if verbose_raw and data_set.suffix:
        pytest.skip("Skipping Verbose mode for this data sets")

    app_client = EthAppClient(backend)
    device = backend.device

    snapshots_dirname = test_name + data_set.suffix
    if verbose_raw:
        settings_toggle(device, navigator, [SettingID.DISPLAY_HASH])
        snapshots_dirname += "-verbose"

    vrs = eip712_new_common(device,
                            navigator,
                            default_screenshot_path,
                            app_client,
                            data_set.data,
                            data_set.filters,
                            golden_run)

    # verify signature
    addr = recover_message(data_set.data, vrs)
    assert addr == get_wallet_addr(app_client)


def test_eip712_filtering_empty_array(backend: BackendInterface,
                                      navigator: Navigator,
                                      default_screenshot_path: Path,
                                      test_name: str,
                                      golden_run: bool,
                                      simu_params: Optional[TxSimu] = None):
    global snapshots_dirname
    global validate_warning

    app_client = EthAppClient(backend)
    device = backend.device

    snapshots_dirname = test_name

    data = {
        "types": {
            "EIP712Domain": [
                {"name": "name", "type": "string"},
                {"name": "version", "type": "string"},
                {"name": "chainId", "type": "uint256"},
                {"name": "verifyingContract", "type": "address"},
            ],
            "Person": [
                {"name": "name", "type": "string"},
                {"name": "addr", "type": "address"},
            ],
            "Message": [
                {"name": "title", "type": "string"},
                {"name": "to", "type": "Person[]"},
            ],
            "Root": [
                {"name": "text", "type": "string"},
                {"name": "subtext", "type": "string[]"},
                {"name": "msg_list1", "type": "Message[]"},
                {"name": "msg_list2", "type": "Message[]"},
            ],
        },
        "primaryType": "Root",
        "domain": {
            "name": "test",
            "version": "1",
            "verifyingContract": "0x0000000000000000000000000000000000000000",
            "chainId": 1,
        },
        "message": {
            "text": "This is a test",
            "subtext": [],
            "msg_list1": [
                {
                    "title": "This is a test",
                    "to": [],
                }
            ],
            "msg_list2": [],
        }
    }
    filters = {
        "name": "Empty array filtering",
        "fields": {
            "text": {
                "type": "raw",
                "name": "Text",
            },
            "subtext.[]": {
                "type": "raw",
                "name": "Sub-Text",
            },
            "msg_list1.[].to.[].addr": {
                "type": "raw",
                "name": "(1) Recipient addr",
            },
            "msg_list2.[].to.[].addr": {
                "type": "raw",
                "name": "(2) Recipient addr",
            },
        }
    }

    if simu_params is not None:
        validate_warning = True
        smsg = encode_typed_data(full_message=data)
        simu_params.tx_hash = smsg.body
        simu_params.domain_hash = smsg.header
        response = app_client.provide_tx_simulation(simu_params)
        assert response.status == StatusWord.OK

    vrs = eip712_new_common(device,
                            navigator,
                            default_screenshot_path,
                            app_client,
                            data,
                            filters,
                            golden_run)

    # verify signature
    addr = recover_message(data, vrs)
    assert addr == get_wallet_addr(app_client)


TOKENS = [
    [
        {
            "addr": "0x1111111111111111111111111111111111111111",
            "ticker": "SRC",
            "decimals": 18,
            "chain_id": 1,
        },
        {},
    ],
    [
        {},
        {
            "addr": "0x2222222222222222222222222222222222222222",
            "ticker": "DST",
            "decimals": 18,
            "chain_id": 1,
        },
    ]
]


@pytest.fixture(name="tokens", params=TOKENS)
def tokens_fixture(request) -> list[dict]:
    return request.param


def test_eip712_advanced_missing_token(backend: BackendInterface,
                                       navigator: Navigator,
                                       default_screenshot_path: Path,
                                       test_name: str,
                                       tokens: list[dict],
                                       golden_run: bool):
    global snapshots_dirname

    test_name += f"-{len(tokens[0]) == 0}-{len(tokens[1]) == 0}"
    snapshots_dirname = test_name

    app_client = EthAppClient(backend)
    device = backend.device

    data = {
        "types": {
            "EIP712Domain": [
                {"name": "name", "type": "string"},
                {"name": "version", "type": "string"},
                {"name": "chainId", "type": "uint256"},
                {"name": "verifyingContract", "type": "address"},
            ],
            "Root": [
                {"name": "token_from", "type": "address"},
                {"name": "value_from", "type": "uint256"},
                {"name": "token_to", "type": "address"},
                {"name": "value_to", "type": "uint256"},
            ]
        },
        "primaryType": "Root",
        "domain": {
            "name": "test",
            "version": "1",
            "verifyingContract": "0x0000000000000000000000000000000000000000",
            "chainId": 1,
        },
        "message": {
            "token_from": "0x1111111111111111111111111111111111111111",
            "value_from": web3.Web3.to_wei(3.65, "ether"),
            "token_to": "0x2222222222222222222222222222222222222222",
            "value_to": web3.Web3.to_wei(15.47, "ether"),
        }
    }

    filters = {
        "name": "Token not in CAL test",
        "tokens": tokens,
        "fields": {
            "token_from": {
                "type": "amount_join_token",
                "token": 0,
            },
            "value_from": {
                "type": "amount_join_value",
                "name": "From",
                "token": 0,
            },
            "token_to": {
                "type": "amount_join_token",
                "token": 1,
            },
            "value_to": {
                "type": "amount_join_value",
                "name": "To",
                "token": 1,
            },
        }
    }
    vrs = eip712_new_common(device,
                            navigator,
                            default_screenshot_path,
                            app_client,
                            data,
                            filters,
                            golden_run)

    # verify signature
    addr = recover_message(data, vrs)
    assert addr == get_wallet_addr(app_client)


TRUSTED_NAMES = [
    (TrustedNameType.CONTRACT, TrustedNameSource.CAL, "Validator contract"),
    (TrustedNameType.ACCOUNT, TrustedNameSource.ENS, "validator.eth"),
]

FILT_TN_TYPES = [
    [TrustedNameType.CONTRACT],
    [TrustedNameType.ACCOUNT],
    [TrustedNameType.CONTRACT, TrustedNameType.ACCOUNT],
    [TrustedNameType.ACCOUNT, TrustedNameType.CONTRACT],
]


@pytest.fixture(name="trusted_name", params=TRUSTED_NAMES)
def trusted_name_fixture(request) -> tuple:
    return request.param


@pytest.fixture(name="filt_tn_types", params=FILT_TN_TYPES)
def filt_tn_types_fixture(request) -> list[TrustedNameType]:
    return request.param


def test_eip712_advanced_trusted_name(backend: BackendInterface,
                                      navigator: Navigator,
                                      default_screenshot_path: Path,
                                      test_name: str,
                                      trusted_name: tuple,
                                      filt_tn_types: list[TrustedNameType],
                                      golden_run: bool):
    global snapshots_dirname

    test_name += f"_{trusted_name[0].name.lower()}_with"
    for t in filt_tn_types:
        test_name += f"_{t.name.lower()}"
    snapshots_dirname = test_name

    app_client = EthAppClient(backend)
    device = backend.device

    data = {
        "types": {
            "EIP712Domain": [
                {"name": "name", "type": "string"},
                {"name": "version", "type": "string"},
                {"name": "chainId", "type": "uint256"},
                {"name": "verifyingContract", "type": "address"},
            ],
            "Root": [
                {"name": "validator", "type": "address"},
                {"name": "enable", "type": "bool"},
            ]
        },
        "primaryType": "Root",
        "domain": {
            "name": "test",
            "version": "1",
            "verifyingContract": "0x0000000000000000000000000000000000000000",
            "chainId": 1,
        },
        "message": {
            "validator": "0x1111111111111111111111111111111111111111",
            "enable": True,
        }
    }

    filters = {
        "name": "Trusted name test",
        "fields": {
            "validator": {
                "type": "trusted_name",
                "name": "Validator",
                "tn_type": filt_tn_types,
                "tn_source": [TrustedNameSource.CAL, TrustedNameSource.ENS],
            },
            "enable": {
                "type": "raw",
                "name": "State",
            },
        }
    }

    if trusted_name[0] is TrustedNameType.ACCOUNT:
        challenge = ResponseParser.challenge(app_client.get_challenge().data)
    else:
        challenge = None

    app_client.provide_trusted_name_v2(bytes.fromhex(data["message"]["validator"][2:]),
                                       trusted_name[2],
                                       trusted_name[0],
                                       trusted_name[1],
                                       data["domain"]["chainId"],
                                       challenge=challenge)
    vrs = eip712_new_common(device,
                            navigator,
                            default_screenshot_path,
                            app_client,
                            data,
                            filters,
                            golden_run)

    # verify signature
    addr = recover_message(data, vrs)
    assert addr == get_wallet_addr(app_client)


def test_eip712_bs_not_activated_error(backend: BackendInterface,
                                       navigator: Navigator,
                                       default_screenshot_path: Path):
    app_client = EthAppClient(backend)
    device = backend.device

    with pytest.raises(ExceptionRAPDU) as e:
        eip712_new_common(device,
                          navigator,
                          default_screenshot_path,
                          app_client,
                          ADVANCED_DATA_SETS[0].data,
                          None,
                          False)
    InputData.disable_autonext()  # so the timer stops firing
    assert e.value.status == StatusWord.INVALID_DATA


def test_eip712_skip(backend: BackendInterface,
                     navigator: Navigator,
                     default_screenshot_path: Path,
                     golden_run: bool):
    global validate_warning
    global skip_flow

    app_client = EthAppClient(backend)
    device = backend.device

    if device.is_nano:
        pytest.skip("Not supported on Nano devices")

    validate_warning = True
    skip_flow = True
    settings_toggle(device, navigator, [SettingID.BLIND_SIGNING])
    with open(input_files()[0], encoding="utf-8") as file:
        data = json.load(file)
    vrs = eip712_new_common(device,
                            navigator,
                            default_screenshot_path,
                            app_client,
                            data,
                            None,
                            golden_run)

    # verify signature
    addr = recover_message(data, vrs)
    assert addr == get_wallet_addr(app_client)


def test_eip712_proxy(backend: BackendInterface,
                      navigator: Navigator,
                      default_screenshot_path: Path):
    app_client = EthAppClient(backend)
    device = backend.device

    input_file = input_files()[0]
    with open(input_file, encoding="utf-8") as file:
        data = json.load(file)
    with open(get_filter_file_from_data_file(Path(input_file)), encoding="utf-8") as file:
        filters = json.load(file)
    # change its name & set a different address than the one in verifyingContract
    filters["name"] = "Proxy test"
    filters["address"] = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

    proxy_info = ProxyInfo(
        ResponseParser.challenge(app_client.get_challenge().data),
        bytes.fromhex(filters["address"][2:]),
        int(data["domain"]["chainId"]),
        bytes.fromhex(data["domain"]["verifyingContract"][2:]),
    )

    app_client.provide_proxy_info(proxy_info.serialize())

    vrs = eip712_new_common(device,
                            navigator,
                            default_screenshot_path,
                            app_client,
                            data,
                            filters,
                            False)

    # verify signature
    addr = recover_message(data, vrs)
    assert addr == get_wallet_addr(app_client)
