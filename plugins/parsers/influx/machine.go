
//line plugins/parsers/influx/machine.go.rl:1
package influx

import (
	"errors"
)

var (
	ErrNameParse = errors.New("expected measurement name")
	ErrFieldParse = errors.New("expected field")
	ErrTagParse = errors.New("expected tag")
	ErrTimestampParse = errors.New("expected timestamp")
	ErrParse = errors.New("parse error")
)


//line plugins/parsers/influx/machine.go.rl:224



//line plugins/parsers/influx/machine.go:23
const LineProtocol_start int = 1
const LineProtocol_first_final int = 191
const LineProtocol_error int = 0

const LineProtocol_en_main int = 1
const LineProtocol_en_discard_line int = 187
const LineProtocol_en_align int = 188


//line plugins/parsers/influx/machine.go.rl:227

type Handler interface {
	SetMeasurement(name []byte)
	AddTag(key []byte, value []byte)
	AddInt(key []byte, value []byte)
	AddUint(key []byte, value []byte)
	AddFloat(key []byte, value []byte)
	AddString(key []byte, value []byte)
	AddBool(key []byte, value []byte)
	SetTimestamp(tm []byte)
}

type machine struct {
	data       []byte
	cs         int
	p, pe, eof int
	pb         int
	handler    Handler
	err        error
}

func NewMachine(handler Handler) *machine {
	m := &machine{
		handler: handler,
	}

	
//line plugins/parsers/influx/machine.go.rl:254
	
//line plugins/parsers/influx/machine.go.rl:255
	
//line plugins/parsers/influx/machine.go.rl:256
	
//line plugins/parsers/influx/machine.go.rl:257
	
//line plugins/parsers/influx/machine.go.rl:258
	
//line plugins/parsers/influx/machine.go:71
	{
	 m.cs = LineProtocol_start
	}

//line plugins/parsers/influx/machine.go.rl:259

	return m
}

func (m *machine) SetData(data []byte) {
	m.data = data
	m.p = 0
	m.pb = 0
	m.pe = len(data)
	m.eof = len(data)
	m.err = nil

	
//line plugins/parsers/influx/machine.go:90
	{
	 m.cs = LineProtocol_start
	}

//line plugins/parsers/influx/machine.go.rl:272
	m.cs = LineProtocol_en_align
}

// ParseLine parses a line of input and returns true if more data can be
// parsed.
func (m *machine) ParseLine() bool {
	if m.data == nil || m.p >= m.pe {
		m.err = nil
		return false
	}

	m.err = nil
	var key []byte
	var yield bool

	
//line plugins/parsers/influx/machine.go:112
	{
	if ( m.p) == ( m.pe) {
		goto _test_eof
	}
	goto _resume

_again:
	switch  m.cs {
	case 1:
		goto st1
	case 2:
		goto st2
	case 3:
		goto st3
	case 4:
		goto st4
	case 0:
		goto st0
	case 5:
		goto st5
	case 6:
		goto st6
	case 7:
		goto st7
	case 191:
		goto st191
	case 192:
		goto st192
	case 193:
		goto st193
	case 8:
		goto st8
	case 194:
		goto st194
	case 195:
		goto st195
	case 196:
		goto st196
	case 197:
		goto st197
	case 198:
		goto st198
	case 199:
		goto st199
	case 200:
		goto st200
	case 201:
		goto st201
	case 202:
		goto st202
	case 203:
		goto st203
	case 204:
		goto st204
	case 205:
		goto st205
	case 206:
		goto st206
	case 207:
		goto st207
	case 208:
		goto st208
	case 209:
		goto st209
	case 210:
		goto st210
	case 211:
		goto st211
	case 212:
		goto st212
	case 213:
		goto st213
	case 9:
		goto st9
	case 10:
		goto st10
	case 11:
		goto st11
	case 12:
		goto st12
	case 214:
		goto st214
	case 215:
		goto st215
	case 13:
		goto st13
	case 14:
		goto st14
	case 216:
		goto st216
	case 217:
		goto st217
	case 218:
		goto st218
	case 219:
		goto st219
	case 220:
		goto st220
	case 221:
		goto st221
	case 222:
		goto st222
	case 15:
		goto st15
	case 16:
		goto st16
	case 17:
		goto st17
	case 223:
		goto st223
	case 18:
		goto st18
	case 19:
		goto st19
	case 20:
		goto st20
	case 224:
		goto st224
	case 21:
		goto st21
	case 22:
		goto st22
	case 225:
		goto st225
	case 226:
		goto st226
	case 23:
		goto st23
	case 24:
		goto st24
	case 25:
		goto st25
	case 26:
		goto st26
	case 27:
		goto st27
	case 28:
		goto st28
	case 29:
		goto st29
	case 30:
		goto st30
	case 31:
		goto st31
	case 32:
		goto st32
	case 33:
		goto st33
	case 34:
		goto st34
	case 35:
		goto st35
	case 36:
		goto st36
	case 37:
		goto st37
	case 38:
		goto st38
	case 39:
		goto st39
	case 40:
		goto st40
	case 41:
		goto st41
	case 227:
		goto st227
	case 228:
		goto st228
	case 42:
		goto st42
	case 229:
		goto st229
	case 230:
		goto st230
	case 231:
		goto st231
	case 232:
		goto st232
	case 233:
		goto st233
	case 234:
		goto st234
	case 235:
		goto st235
	case 236:
		goto st236
	case 237:
		goto st237
	case 238:
		goto st238
	case 239:
		goto st239
	case 240:
		goto st240
	case 241:
		goto st241
	case 242:
		goto st242
	case 243:
		goto st243
	case 244:
		goto st244
	case 245:
		goto st245
	case 246:
		goto st246
	case 247:
		goto st247
	case 248:
		goto st248
	case 43:
		goto st43
	case 249:
		goto st249
	case 250:
		goto st250
	case 44:
		goto st44
	case 251:
		goto st251
	case 252:
		goto st252
	case 253:
		goto st253
	case 254:
		goto st254
	case 255:
		goto st255
	case 256:
		goto st256
	case 257:
		goto st257
	case 258:
		goto st258
	case 259:
		goto st259
	case 260:
		goto st260
	case 261:
		goto st261
	case 262:
		goto st262
	case 263:
		goto st263
	case 264:
		goto st264
	case 265:
		goto st265
	case 266:
		goto st266
	case 267:
		goto st267
	case 268:
		goto st268
	case 269:
		goto st269
	case 270:
		goto st270
	case 45:
		goto st45
	case 46:
		goto st46
	case 47:
		goto st47
	case 271:
		goto st271
	case 48:
		goto st48
	case 49:
		goto st49
	case 50:
		goto st50
	case 51:
		goto st51
	case 272:
		goto st272
	case 273:
		goto st273
	case 52:
		goto st52
	case 274:
		goto st274
	case 53:
		goto st53
	case 275:
		goto st275
	case 276:
		goto st276
	case 277:
		goto st277
	case 278:
		goto st278
	case 279:
		goto st279
	case 280:
		goto st280
	case 281:
		goto st281
	case 54:
		goto st54
	case 55:
		goto st55
	case 56:
		goto st56
	case 282:
		goto st282
	case 57:
		goto st57
	case 58:
		goto st58
	case 59:
		goto st59
	case 283:
		goto st283
	case 60:
		goto st60
	case 61:
		goto st61
	case 284:
		goto st284
	case 285:
		goto st285
	case 62:
		goto st62
	case 63:
		goto st63
	case 286:
		goto st286
	case 287:
		goto st287
	case 64:
		goto st64
	case 65:
		goto st65
	case 288:
		goto st288
	case 289:
		goto st289
	case 290:
		goto st290
	case 291:
		goto st291
	case 292:
		goto st292
	case 293:
		goto st293
	case 294:
		goto st294
	case 66:
		goto st66
	case 67:
		goto st67
	case 68:
		goto st68
	case 295:
		goto st295
	case 69:
		goto st69
	case 70:
		goto st70
	case 71:
		goto st71
	case 296:
		goto st296
	case 72:
		goto st72
	case 73:
		goto st73
	case 297:
		goto st297
	case 298:
		goto st298
	case 74:
		goto st74
	case 75:
		goto st75
	case 76:
		goto st76
	case 77:
		goto st77
	case 78:
		goto st78
	case 79:
		goto st79
	case 299:
		goto st299
	case 300:
		goto st300
	case 301:
		goto st301
	case 302:
		goto st302
	case 80:
		goto st80
	case 303:
		goto st303
	case 304:
		goto st304
	case 305:
		goto st305
	case 306:
		goto st306
	case 81:
		goto st81
	case 307:
		goto st307
	case 308:
		goto st308
	case 309:
		goto st309
	case 310:
		goto st310
	case 311:
		goto st311
	case 312:
		goto st312
	case 313:
		goto st313
	case 314:
		goto st314
	case 315:
		goto st315
	case 316:
		goto st316
	case 317:
		goto st317
	case 318:
		goto st318
	case 319:
		goto st319
	case 320:
		goto st320
	case 321:
		goto st321
	case 322:
		goto st322
	case 323:
		goto st323
	case 324:
		goto st324
	case 82:
		goto st82
	case 83:
		goto st83
	case 84:
		goto st84
	case 85:
		goto st85
	case 86:
		goto st86
	case 87:
		goto st87
	case 88:
		goto st88
	case 89:
		goto st89
	case 90:
		goto st90
	case 91:
		goto st91
	case 92:
		goto st92
	case 93:
		goto st93
	case 94:
		goto st94
	case 325:
		goto st325
	case 326:
		goto st326
	case 95:
		goto st95
	case 327:
		goto st327
	case 328:
		goto st328
	case 329:
		goto st329
	case 330:
		goto st330
	case 331:
		goto st331
	case 332:
		goto st332
	case 333:
		goto st333
	case 334:
		goto st334
	case 335:
		goto st335
	case 336:
		goto st336
	case 337:
		goto st337
	case 338:
		goto st338
	case 339:
		goto st339
	case 340:
		goto st340
	case 341:
		goto st341
	case 342:
		goto st342
	case 343:
		goto st343
	case 344:
		goto st344
	case 345:
		goto st345
	case 346:
		goto st346
	case 96:
		goto st96
	case 97:
		goto st97
	case 347:
		goto st347
	case 348:
		goto st348
	case 98:
		goto st98
	case 349:
		goto st349
	case 350:
		goto st350
	case 351:
		goto st351
	case 352:
		goto st352
	case 353:
		goto st353
	case 354:
		goto st354
	case 355:
		goto st355
	case 356:
		goto st356
	case 357:
		goto st357
	case 358:
		goto st358
	case 359:
		goto st359
	case 360:
		goto st360
	case 361:
		goto st361
	case 362:
		goto st362
	case 363:
		goto st363
	case 364:
		goto st364
	case 365:
		goto st365
	case 366:
		goto st366
	case 367:
		goto st367
	case 368:
		goto st368
	case 99:
		goto st99
	case 369:
		goto st369
	case 370:
		goto st370
	case 100:
		goto st100
	case 101:
		goto st101
	case 102:
		goto st102
	case 103:
		goto st103
	case 371:
		goto st371
	case 372:
		goto st372
	case 104:
		goto st104
	case 105:
		goto st105
	case 373:
		goto st373
	case 374:
		goto st374
	case 375:
		goto st375
	case 376:
		goto st376
	case 377:
		goto st377
	case 378:
		goto st378
	case 379:
		goto st379
	case 106:
		goto st106
	case 107:
		goto st107
	case 108:
		goto st108
	case 380:
		goto st380
	case 109:
		goto st109
	case 110:
		goto st110
	case 111:
		goto st111
	case 381:
		goto st381
	case 112:
		goto st112
	case 113:
		goto st113
	case 382:
		goto st382
	case 383:
		goto st383
	case 114:
		goto st114
	case 115:
		goto st115
	case 116:
		goto st116
	case 117:
		goto st117
	case 118:
		goto st118
	case 119:
		goto st119
	case 120:
		goto st120
	case 121:
		goto st121
	case 122:
		goto st122
	case 123:
		goto st123
	case 124:
		goto st124
	case 125:
		goto st125
	case 384:
		goto st384
	case 385:
		goto st385
	case 386:
		goto st386
	case 126:
		goto st126
	case 387:
		goto st387
	case 388:
		goto st388
	case 389:
		goto st389
	case 390:
		goto st390
	case 391:
		goto st391
	case 392:
		goto st392
	case 393:
		goto st393
	case 394:
		goto st394
	case 395:
		goto st395
	case 396:
		goto st396
	case 397:
		goto st397
	case 398:
		goto st398
	case 399:
		goto st399
	case 400:
		goto st400
	case 401:
		goto st401
	case 402:
		goto st402
	case 403:
		goto st403
	case 404:
		goto st404
	case 405:
		goto st405
	case 406:
		goto st406
	case 407:
		goto st407
	case 408:
		goto st408
	case 127:
		goto st127
	case 409:
		goto st409
	case 410:
		goto st410
	case 411:
		goto st411
	case 412:
		goto st412
	case 128:
		goto st128
	case 413:
		goto st413
	case 414:
		goto st414
	case 415:
		goto st415
	case 416:
		goto st416
	case 417:
		goto st417
	case 418:
		goto st418
	case 419:
		goto st419
	case 420:
		goto st420
	case 421:
		goto st421
	case 422:
		goto st422
	case 423:
		goto st423
	case 424:
		goto st424
	case 425:
		goto st425
	case 426:
		goto st426
	case 427:
		goto st427
	case 428:
		goto st428
	case 429:
		goto st429
	case 430:
		goto st430
	case 431:
		goto st431
	case 432:
		goto st432
	case 129:
		goto st129
	case 130:
		goto st130
	case 131:
		goto st131
	case 433:
		goto st433
	case 434:
		goto st434
	case 132:
		goto st132
	case 435:
		goto st435
	case 436:
		goto st436
	case 437:
		goto st437
	case 438:
		goto st438
	case 439:
		goto st439
	case 440:
		goto st440
	case 441:
		goto st441
	case 442:
		goto st442
	case 443:
		goto st443
	case 444:
		goto st444
	case 445:
		goto st445
	case 446:
		goto st446
	case 447:
		goto st447
	case 448:
		goto st448
	case 449:
		goto st449
	case 450:
		goto st450
	case 451:
		goto st451
	case 452:
		goto st452
	case 453:
		goto st453
	case 454:
		goto st454
	case 133:
		goto st133
	case 455:
		goto st455
	case 456:
		goto st456
	case 457:
		goto st457
	case 134:
		goto st134
	case 458:
		goto st458
	case 459:
		goto st459
	case 460:
		goto st460
	case 461:
		goto st461
	case 462:
		goto st462
	case 463:
		goto st463
	case 464:
		goto st464
	case 465:
		goto st465
	case 466:
		goto st466
	case 467:
		goto st467
	case 468:
		goto st468
	case 469:
		goto st469
	case 470:
		goto st470
	case 471:
		goto st471
	case 472:
		goto st472
	case 473:
		goto st473
	case 474:
		goto st474
	case 475:
		goto st475
	case 476:
		goto st476
	case 477:
		goto st477
	case 478:
		goto st478
	case 479:
		goto st479
	case 135:
		goto st135
	case 480:
		goto st480
	case 481:
		goto st481
	case 482:
		goto st482
	case 483:
		goto st483
	case 484:
		goto st484
	case 485:
		goto st485
	case 486:
		goto st486
	case 487:
		goto st487
	case 488:
		goto st488
	case 489:
		goto st489
	case 490:
		goto st490
	case 491:
		goto st491
	case 492:
		goto st492
	case 493:
		goto st493
	case 494:
		goto st494
	case 495:
		goto st495
	case 496:
		goto st496
	case 497:
		goto st497
	case 498:
		goto st498
	case 499:
		goto st499
	case 500:
		goto st500
	case 501:
		goto st501
	case 136:
		goto st136
	case 137:
		goto st137
	case 138:
		goto st138
	case 139:
		goto st139
	case 502:
		goto st502
	case 503:
		goto st503
	case 140:
		goto st140
	case 504:
		goto st504
	case 141:
		goto st141
	case 505:
		goto st505
	case 506:
		goto st506
	case 507:
		goto st507
	case 508:
		goto st508
	case 509:
		goto st509
	case 510:
		goto st510
	case 511:
		goto st511
	case 142:
		goto st142
	case 143:
		goto st143
	case 144:
		goto st144
	case 512:
		goto st512
	case 145:
		goto st145
	case 146:
		goto st146
	case 147:
		goto st147
	case 513:
		goto st513
	case 148:
		goto st148
	case 149:
		goto st149
	case 514:
		goto st514
	case 515:
		goto st515
	case 150:
		goto st150
	case 151:
		goto st151
	case 516:
		goto st516
	case 517:
		goto st517
	case 518:
		goto st518
	case 152:
		goto st152
	case 519:
		goto st519
	case 520:
		goto st520
	case 521:
		goto st521
	case 522:
		goto st522
	case 523:
		goto st523
	case 524:
		goto st524
	case 525:
		goto st525
	case 526:
		goto st526
	case 527:
		goto st527
	case 528:
		goto st528
	case 529:
		goto st529
	case 530:
		goto st530
	case 531:
		goto st531
	case 532:
		goto st532
	case 533:
		goto st533
	case 534:
		goto st534
	case 535:
		goto st535
	case 536:
		goto st536
	case 537:
		goto st537
	case 538:
		goto st538
	case 539:
		goto st539
	case 153:
		goto st153
	case 154:
		goto st154
	case 540:
		goto st540
	case 541:
		goto st541
	case 542:
		goto st542
	case 543:
		goto st543
	case 544:
		goto st544
	case 545:
		goto st545
	case 546:
		goto st546
	case 155:
		goto st155
	case 156:
		goto st156
	case 157:
		goto st157
	case 547:
		goto st547
	case 158:
		goto st158
	case 159:
		goto st159
	case 160:
		goto st160
	case 548:
		goto st548
	case 161:
		goto st161
	case 162:
		goto st162
	case 549:
		goto st549
	case 550:
		goto st550
	case 163:
		goto st163
	case 164:
		goto st164
	case 165:
		goto st165
	case 551:
		goto st551
	case 552:
		goto st552
	case 166:
		goto st166
	case 553:
		goto st553
	case 554:
		goto st554
	case 167:
		goto st167
	case 555:
		goto st555
	case 556:
		goto st556
	case 557:
		goto st557
	case 558:
		goto st558
	case 559:
		goto st559
	case 560:
		goto st560
	case 561:
		goto st561
	case 168:
		goto st168
	case 169:
		goto st169
	case 170:
		goto st170
	case 562:
		goto st562
	case 171:
		goto st171
	case 172:
		goto st172
	case 173:
		goto st173
	case 563:
		goto st563
	case 174:
		goto st174
	case 175:
		goto st175
	case 564:
		goto st564
	case 565:
		goto st565
	case 176:
		goto st176
	case 566:
		goto st566
	case 567:
		goto st567
	case 177:
		goto st177
	case 178:
		goto st178
	case 568:
		goto st568
	case 569:
		goto st569
	case 570:
		goto st570
	case 571:
		goto st571
	case 572:
		goto st572
	case 573:
		goto st573
	case 179:
		goto st179
	case 180:
		goto st180
	case 181:
		goto st181
	case 574:
		goto st574
	case 182:
		goto st182
	case 183:
		goto st183
	case 184:
		goto st184
	case 575:
		goto st575
	case 185:
		goto st185
	case 186:
		goto st186
	case 576:
		goto st576
	case 577:
		goto st577
	case 187:
		goto st187
	case 578:
		goto st578
	case 188:
		goto st188
	case 579:
		goto st579
	case 580:
		goto st580
	case 189:
		goto st189
	case 190:
		goto st190
	}

	if ( m.p)++; ( m.p) == ( m.pe) {
		goto _test_eof
	}
_resume:
	switch  m.cs {
	case 1:
		goto st_case_1
	case 2:
		goto st_case_2
	case 3:
		goto st_case_3
	case 4:
		goto st_case_4
	case 0:
		goto st_case_0
	case 5:
		goto st_case_5
	case 6:
		goto st_case_6
	case 7:
		goto st_case_7
	case 191:
		goto st_case_191
	case 192:
		goto st_case_192
	case 193:
		goto st_case_193
	case 8:
		goto st_case_8
	case 194:
		goto st_case_194
	case 195:
		goto st_case_195
	case 196:
		goto st_case_196
	case 197:
		goto st_case_197
	case 198:
		goto st_case_198
	case 199:
		goto st_case_199
	case 200:
		goto st_case_200
	case 201:
		goto st_case_201
	case 202:
		goto st_case_202
	case 203:
		goto st_case_203
	case 204:
		goto st_case_204
	case 205:
		goto st_case_205
	case 206:
		goto st_case_206
	case 207:
		goto st_case_207
	case 208:
		goto st_case_208
	case 209:
		goto st_case_209
	case 210:
		goto st_case_210
	case 211:
		goto st_case_211
	case 212:
		goto st_case_212
	case 213:
		goto st_case_213
	case 9:
		goto st_case_9
	case 10:
		goto st_case_10
	case 11:
		goto st_case_11
	case 12:
		goto st_case_12
	case 214:
		goto st_case_214
	case 215:
		goto st_case_215
	case 13:
		goto st_case_13
	case 14:
		goto st_case_14
	case 216:
		goto st_case_216
	case 217:
		goto st_case_217
	case 218:
		goto st_case_218
	case 219:
		goto st_case_219
	case 220:
		goto st_case_220
	case 221:
		goto st_case_221
	case 222:
		goto st_case_222
	case 15:
		goto st_case_15
	case 16:
		goto st_case_16
	case 17:
		goto st_case_17
	case 223:
		goto st_case_223
	case 18:
		goto st_case_18
	case 19:
		goto st_case_19
	case 20:
		goto st_case_20
	case 224:
		goto st_case_224
	case 21:
		goto st_case_21
	case 22:
		goto st_case_22
	case 225:
		goto st_case_225
	case 226:
		goto st_case_226
	case 23:
		goto st_case_23
	case 24:
		goto st_case_24
	case 25:
		goto st_case_25
	case 26:
		goto st_case_26
	case 27:
		goto st_case_27
	case 28:
		goto st_case_28
	case 29:
		goto st_case_29
	case 30:
		goto st_case_30
	case 31:
		goto st_case_31
	case 32:
		goto st_case_32
	case 33:
		goto st_case_33
	case 34:
		goto st_case_34
	case 35:
		goto st_case_35
	case 36:
		goto st_case_36
	case 37:
		goto st_case_37
	case 38:
		goto st_case_38
	case 39:
		goto st_case_39
	case 40:
		goto st_case_40
	case 41:
		goto st_case_41
	case 227:
		goto st_case_227
	case 228:
		goto st_case_228
	case 42:
		goto st_case_42
	case 229:
		goto st_case_229
	case 230:
		goto st_case_230
	case 231:
		goto st_case_231
	case 232:
		goto st_case_232
	case 233:
		goto st_case_233
	case 234:
		goto st_case_234
	case 235:
		goto st_case_235
	case 236:
		goto st_case_236
	case 237:
		goto st_case_237
	case 238:
		goto st_case_238
	case 239:
		goto st_case_239
	case 240:
		goto st_case_240
	case 241:
		goto st_case_241
	case 242:
		goto st_case_242
	case 243:
		goto st_case_243
	case 244:
		goto st_case_244
	case 245:
		goto st_case_245
	case 246:
		goto st_case_246
	case 247:
		goto st_case_247
	case 248:
		goto st_case_248
	case 43:
		goto st_case_43
	case 249:
		goto st_case_249
	case 250:
		goto st_case_250
	case 44:
		goto st_case_44
	case 251:
		goto st_case_251
	case 252:
		goto st_case_252
	case 253:
		goto st_case_253
	case 254:
		goto st_case_254
	case 255:
		goto st_case_255
	case 256:
		goto st_case_256
	case 257:
		goto st_case_257
	case 258:
		goto st_case_258
	case 259:
		goto st_case_259
	case 260:
		goto st_case_260
	case 261:
		goto st_case_261
	case 262:
		goto st_case_262
	case 263:
		goto st_case_263
	case 264:
		goto st_case_264
	case 265:
		goto st_case_265
	case 266:
		goto st_case_266
	case 267:
		goto st_case_267
	case 268:
		goto st_case_268
	case 269:
		goto st_case_269
	case 270:
		goto st_case_270
	case 45:
		goto st_case_45
	case 46:
		goto st_case_46
	case 47:
		goto st_case_47
	case 271:
		goto st_case_271
	case 48:
		goto st_case_48
	case 49:
		goto st_case_49
	case 50:
		goto st_case_50
	case 51:
		goto st_case_51
	case 272:
		goto st_case_272
	case 273:
		goto st_case_273
	case 52:
		goto st_case_52
	case 274:
		goto st_case_274
	case 53:
		goto st_case_53
	case 275:
		goto st_case_275
	case 276:
		goto st_case_276
	case 277:
		goto st_case_277
	case 278:
		goto st_case_278
	case 279:
		goto st_case_279
	case 280:
		goto st_case_280
	case 281:
		goto st_case_281
	case 54:
		goto st_case_54
	case 55:
		goto st_case_55
	case 56:
		goto st_case_56
	case 282:
		goto st_case_282
	case 57:
		goto st_case_57
	case 58:
		goto st_case_58
	case 59:
		goto st_case_59
	case 283:
		goto st_case_283
	case 60:
		goto st_case_60
	case 61:
		goto st_case_61
	case 284:
		goto st_case_284
	case 285:
		goto st_case_285
	case 62:
		goto st_case_62
	case 63:
		goto st_case_63
	case 286:
		goto st_case_286
	case 287:
		goto st_case_287
	case 64:
		goto st_case_64
	case 65:
		goto st_case_65
	case 288:
		goto st_case_288
	case 289:
		goto st_case_289
	case 290:
		goto st_case_290
	case 291:
		goto st_case_291
	case 292:
		goto st_case_292
	case 293:
		goto st_case_293
	case 294:
		goto st_case_294
	case 66:
		goto st_case_66
	case 67:
		goto st_case_67
	case 68:
		goto st_case_68
	case 295:
		goto st_case_295
	case 69:
		goto st_case_69
	case 70:
		goto st_case_70
	case 71:
		goto st_case_71
	case 296:
		goto st_case_296
	case 72:
		goto st_case_72
	case 73:
		goto st_case_73
	case 297:
		goto st_case_297
	case 298:
		goto st_case_298
	case 74:
		goto st_case_74
	case 75:
		goto st_case_75
	case 76:
		goto st_case_76
	case 77:
		goto st_case_77
	case 78:
		goto st_case_78
	case 79:
		goto st_case_79
	case 299:
		goto st_case_299
	case 300:
		goto st_case_300
	case 301:
		goto st_case_301
	case 302:
		goto st_case_302
	case 80:
		goto st_case_80
	case 303:
		goto st_case_303
	case 304:
		goto st_case_304
	case 305:
		goto st_case_305
	case 306:
		goto st_case_306
	case 81:
		goto st_case_81
	case 307:
		goto st_case_307
	case 308:
		goto st_case_308
	case 309:
		goto st_case_309
	case 310:
		goto st_case_310
	case 311:
		goto st_case_311
	case 312:
		goto st_case_312
	case 313:
		goto st_case_313
	case 314:
		goto st_case_314
	case 315:
		goto st_case_315
	case 316:
		goto st_case_316
	case 317:
		goto st_case_317
	case 318:
		goto st_case_318
	case 319:
		goto st_case_319
	case 320:
		goto st_case_320
	case 321:
		goto st_case_321
	case 322:
		goto st_case_322
	case 323:
		goto st_case_323
	case 324:
		goto st_case_324
	case 82:
		goto st_case_82
	case 83:
		goto st_case_83
	case 84:
		goto st_case_84
	case 85:
		goto st_case_85
	case 86:
		goto st_case_86
	case 87:
		goto st_case_87
	case 88:
		goto st_case_88
	case 89:
		goto st_case_89
	case 90:
		goto st_case_90
	case 91:
		goto st_case_91
	case 92:
		goto st_case_92
	case 93:
		goto st_case_93
	case 94:
		goto st_case_94
	case 325:
		goto st_case_325
	case 326:
		goto st_case_326
	case 95:
		goto st_case_95
	case 327:
		goto st_case_327
	case 328:
		goto st_case_328
	case 329:
		goto st_case_329
	case 330:
		goto st_case_330
	case 331:
		goto st_case_331
	case 332:
		goto st_case_332
	case 333:
		goto st_case_333
	case 334:
		goto st_case_334
	case 335:
		goto st_case_335
	case 336:
		goto st_case_336
	case 337:
		goto st_case_337
	case 338:
		goto st_case_338
	case 339:
		goto st_case_339
	case 340:
		goto st_case_340
	case 341:
		goto st_case_341
	case 342:
		goto st_case_342
	case 343:
		goto st_case_343
	case 344:
		goto st_case_344
	case 345:
		goto st_case_345
	case 346:
		goto st_case_346
	case 96:
		goto st_case_96
	case 97:
		goto st_case_97
	case 347:
		goto st_case_347
	case 348:
		goto st_case_348
	case 98:
		goto st_case_98
	case 349:
		goto st_case_349
	case 350:
		goto st_case_350
	case 351:
		goto st_case_351
	case 352:
		goto st_case_352
	case 353:
		goto st_case_353
	case 354:
		goto st_case_354
	case 355:
		goto st_case_355
	case 356:
		goto st_case_356
	case 357:
		goto st_case_357
	case 358:
		goto st_case_358
	case 359:
		goto st_case_359
	case 360:
		goto st_case_360
	case 361:
		goto st_case_361
	case 362:
		goto st_case_362
	case 363:
		goto st_case_363
	case 364:
		goto st_case_364
	case 365:
		goto st_case_365
	case 366:
		goto st_case_366
	case 367:
		goto st_case_367
	case 368:
		goto st_case_368
	case 99:
		goto st_case_99
	case 369:
		goto st_case_369
	case 370:
		goto st_case_370
	case 100:
		goto st_case_100
	case 101:
		goto st_case_101
	case 102:
		goto st_case_102
	case 103:
		goto st_case_103
	case 371:
		goto st_case_371
	case 372:
		goto st_case_372
	case 104:
		goto st_case_104
	case 105:
		goto st_case_105
	case 373:
		goto st_case_373
	case 374:
		goto st_case_374
	case 375:
		goto st_case_375
	case 376:
		goto st_case_376
	case 377:
		goto st_case_377
	case 378:
		goto st_case_378
	case 379:
		goto st_case_379
	case 106:
		goto st_case_106
	case 107:
		goto st_case_107
	case 108:
		goto st_case_108
	case 380:
		goto st_case_380
	case 109:
		goto st_case_109
	case 110:
		goto st_case_110
	case 111:
		goto st_case_111
	case 381:
		goto st_case_381
	case 112:
		goto st_case_112
	case 113:
		goto st_case_113
	case 382:
		goto st_case_382
	case 383:
		goto st_case_383
	case 114:
		goto st_case_114
	case 115:
		goto st_case_115
	case 116:
		goto st_case_116
	case 117:
		goto st_case_117
	case 118:
		goto st_case_118
	case 119:
		goto st_case_119
	case 120:
		goto st_case_120
	case 121:
		goto st_case_121
	case 122:
		goto st_case_122
	case 123:
		goto st_case_123
	case 124:
		goto st_case_124
	case 125:
		goto st_case_125
	case 384:
		goto st_case_384
	case 385:
		goto st_case_385
	case 386:
		goto st_case_386
	case 126:
		goto st_case_126
	case 387:
		goto st_case_387
	case 388:
		goto st_case_388
	case 389:
		goto st_case_389
	case 390:
		goto st_case_390
	case 391:
		goto st_case_391
	case 392:
		goto st_case_392
	case 393:
		goto st_case_393
	case 394:
		goto st_case_394
	case 395:
		goto st_case_395
	case 396:
		goto st_case_396
	case 397:
		goto st_case_397
	case 398:
		goto st_case_398
	case 399:
		goto st_case_399
	case 400:
		goto st_case_400
	case 401:
		goto st_case_401
	case 402:
		goto st_case_402
	case 403:
		goto st_case_403
	case 404:
		goto st_case_404
	case 405:
		goto st_case_405
	case 406:
		goto st_case_406
	case 407:
		goto st_case_407
	case 408:
		goto st_case_408
	case 127:
		goto st_case_127
	case 409:
		goto st_case_409
	case 410:
		goto st_case_410
	case 411:
		goto st_case_411
	case 412:
		goto st_case_412
	case 128:
		goto st_case_128
	case 413:
		goto st_case_413
	case 414:
		goto st_case_414
	case 415:
		goto st_case_415
	case 416:
		goto st_case_416
	case 417:
		goto st_case_417
	case 418:
		goto st_case_418
	case 419:
		goto st_case_419
	case 420:
		goto st_case_420
	case 421:
		goto st_case_421
	case 422:
		goto st_case_422
	case 423:
		goto st_case_423
	case 424:
		goto st_case_424
	case 425:
		goto st_case_425
	case 426:
		goto st_case_426
	case 427:
		goto st_case_427
	case 428:
		goto st_case_428
	case 429:
		goto st_case_429
	case 430:
		goto st_case_430
	case 431:
		goto st_case_431
	case 432:
		goto st_case_432
	case 129:
		goto st_case_129
	case 130:
		goto st_case_130
	case 131:
		goto st_case_131
	case 433:
		goto st_case_433
	case 434:
		goto st_case_434
	case 132:
		goto st_case_132
	case 435:
		goto st_case_435
	case 436:
		goto st_case_436
	case 437:
		goto st_case_437
	case 438:
		goto st_case_438
	case 439:
		goto st_case_439
	case 440:
		goto st_case_440
	case 441:
		goto st_case_441
	case 442:
		goto st_case_442
	case 443:
		goto st_case_443
	case 444:
		goto st_case_444
	case 445:
		goto st_case_445
	case 446:
		goto st_case_446
	case 447:
		goto st_case_447
	case 448:
		goto st_case_448
	case 449:
		goto st_case_449
	case 450:
		goto st_case_450
	case 451:
		goto st_case_451
	case 452:
		goto st_case_452
	case 453:
		goto st_case_453
	case 454:
		goto st_case_454
	case 133:
		goto st_case_133
	case 455:
		goto st_case_455
	case 456:
		goto st_case_456
	case 457:
		goto st_case_457
	case 134:
		goto st_case_134
	case 458:
		goto st_case_458
	case 459:
		goto st_case_459
	case 460:
		goto st_case_460
	case 461:
		goto st_case_461
	case 462:
		goto st_case_462
	case 463:
		goto st_case_463
	case 464:
		goto st_case_464
	case 465:
		goto st_case_465
	case 466:
		goto st_case_466
	case 467:
		goto st_case_467
	case 468:
		goto st_case_468
	case 469:
		goto st_case_469
	case 470:
		goto st_case_470
	case 471:
		goto st_case_471
	case 472:
		goto st_case_472
	case 473:
		goto st_case_473
	case 474:
		goto st_case_474
	case 475:
		goto st_case_475
	case 476:
		goto st_case_476
	case 477:
		goto st_case_477
	case 478:
		goto st_case_478
	case 479:
		goto st_case_479
	case 135:
		goto st_case_135
	case 480:
		goto st_case_480
	case 481:
		goto st_case_481
	case 482:
		goto st_case_482
	case 483:
		goto st_case_483
	case 484:
		goto st_case_484
	case 485:
		goto st_case_485
	case 486:
		goto st_case_486
	case 487:
		goto st_case_487
	case 488:
		goto st_case_488
	case 489:
		goto st_case_489
	case 490:
		goto st_case_490
	case 491:
		goto st_case_491
	case 492:
		goto st_case_492
	case 493:
		goto st_case_493
	case 494:
		goto st_case_494
	case 495:
		goto st_case_495
	case 496:
		goto st_case_496
	case 497:
		goto st_case_497
	case 498:
		goto st_case_498
	case 499:
		goto st_case_499
	case 500:
		goto st_case_500
	case 501:
		goto st_case_501
	case 136:
		goto st_case_136
	case 137:
		goto st_case_137
	case 138:
		goto st_case_138
	case 139:
		goto st_case_139
	case 502:
		goto st_case_502
	case 503:
		goto st_case_503
	case 140:
		goto st_case_140
	case 504:
		goto st_case_504
	case 141:
		goto st_case_141
	case 505:
		goto st_case_505
	case 506:
		goto st_case_506
	case 507:
		goto st_case_507
	case 508:
		goto st_case_508
	case 509:
		goto st_case_509
	case 510:
		goto st_case_510
	case 511:
		goto st_case_511
	case 142:
		goto st_case_142
	case 143:
		goto st_case_143
	case 144:
		goto st_case_144
	case 512:
		goto st_case_512
	case 145:
		goto st_case_145
	case 146:
		goto st_case_146
	case 147:
		goto st_case_147
	case 513:
		goto st_case_513
	case 148:
		goto st_case_148
	case 149:
		goto st_case_149
	case 514:
		goto st_case_514
	case 515:
		goto st_case_515
	case 150:
		goto st_case_150
	case 151:
		goto st_case_151
	case 516:
		goto st_case_516
	case 517:
		goto st_case_517
	case 518:
		goto st_case_518
	case 152:
		goto st_case_152
	case 519:
		goto st_case_519
	case 520:
		goto st_case_520
	case 521:
		goto st_case_521
	case 522:
		goto st_case_522
	case 523:
		goto st_case_523
	case 524:
		goto st_case_524
	case 525:
		goto st_case_525
	case 526:
		goto st_case_526
	case 527:
		goto st_case_527
	case 528:
		goto st_case_528
	case 529:
		goto st_case_529
	case 530:
		goto st_case_530
	case 531:
		goto st_case_531
	case 532:
		goto st_case_532
	case 533:
		goto st_case_533
	case 534:
		goto st_case_534
	case 535:
		goto st_case_535
	case 536:
		goto st_case_536
	case 537:
		goto st_case_537
	case 538:
		goto st_case_538
	case 539:
		goto st_case_539
	case 153:
		goto st_case_153
	case 154:
		goto st_case_154
	case 540:
		goto st_case_540
	case 541:
		goto st_case_541
	case 542:
		goto st_case_542
	case 543:
		goto st_case_543
	case 544:
		goto st_case_544
	case 545:
		goto st_case_545
	case 546:
		goto st_case_546
	case 155:
		goto st_case_155
	case 156:
		goto st_case_156
	case 157:
		goto st_case_157
	case 547:
		goto st_case_547
	case 158:
		goto st_case_158
	case 159:
		goto st_case_159
	case 160:
		goto st_case_160
	case 548:
		goto st_case_548
	case 161:
		goto st_case_161
	case 162:
		goto st_case_162
	case 549:
		goto st_case_549
	case 550:
		goto st_case_550
	case 163:
		goto st_case_163
	case 164:
		goto st_case_164
	case 165:
		goto st_case_165
	case 551:
		goto st_case_551
	case 552:
		goto st_case_552
	case 166:
		goto st_case_166
	case 553:
		goto st_case_553
	case 554:
		goto st_case_554
	case 167:
		goto st_case_167
	case 555:
		goto st_case_555
	case 556:
		goto st_case_556
	case 557:
		goto st_case_557
	case 558:
		goto st_case_558
	case 559:
		goto st_case_559
	case 560:
		goto st_case_560
	case 561:
		goto st_case_561
	case 168:
		goto st_case_168
	case 169:
		goto st_case_169
	case 170:
		goto st_case_170
	case 562:
		goto st_case_562
	case 171:
		goto st_case_171
	case 172:
		goto st_case_172
	case 173:
		goto st_case_173
	case 563:
		goto st_case_563
	case 174:
		goto st_case_174
	case 175:
		goto st_case_175
	case 564:
		goto st_case_564
	case 565:
		goto st_case_565
	case 176:
		goto st_case_176
	case 566:
		goto st_case_566
	case 567:
		goto st_case_567
	case 177:
		goto st_case_177
	case 178:
		goto st_case_178
	case 568:
		goto st_case_568
	case 569:
		goto st_case_569
	case 570:
		goto st_case_570
	case 571:
		goto st_case_571
	case 572:
		goto st_case_572
	case 573:
		goto st_case_573
	case 179:
		goto st_case_179
	case 180:
		goto st_case_180
	case 181:
		goto st_case_181
	case 574:
		goto st_case_574
	case 182:
		goto st_case_182
	case 183:
		goto st_case_183
	case 184:
		goto st_case_184
	case 575:
		goto st_case_575
	case 185:
		goto st_case_185
	case 186:
		goto st_case_186
	case 576:
		goto st_case_576
	case 577:
		goto st_case_577
	case 187:
		goto st_case_187
	case 578:
		goto st_case_578
	case 188:
		goto st_case_188
	case 579:
		goto st_case_579
	case 580:
		goto st_case_580
	case 189:
		goto st_case_189
	case 190:
		goto st_case_190
	}
	goto st_out
	st1:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof1
		}
	st_case_1:
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr1
		case 35:
			goto tr1
		case 44:
			goto tr1
		case 92:
			goto tr2
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr1
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr1
		}
		goto tr0
tr0:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st2
	st2:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof2
		}
	st_case_2:
//line plugins/parsers/influx/machine.go:2489
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 11:
			goto tr6
		case 13:
			goto tr5
		case 32:
			goto tr4
		case 44:
			goto tr7
		case 92:
			goto st129
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr4
		}
		goto st2
tr4:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

	goto st3
tr58:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st3
	st3:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof3
		}
	st_case_3:
//line plugins/parsers/influx/machine.go:2525
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 11:
			goto tr11
		case 13:
			goto tr5
		case 32:
			goto st3
		case 44:
			goto tr5
		case 61:
			goto tr5
		case 92:
			goto tr12
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st3
		}
		goto tr9
tr9:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st4
	st4:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof4
		}
	st_case_4:
//line plugins/parsers/influx/machine.go:2557
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr5
		case 44:
			goto tr5
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr5
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr5
		}
		goto st4
tr1:
	 m.cs = 0
//line plugins/parsers/influx/machine.go.rl:56

	m.err = ErrParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++; goto _out }

	goto _again
tr5:
	 m.cs = 0
//line plugins/parsers/influx/machine.go.rl:35

	m.err = ErrFieldParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:56

	m.err = ErrParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++; goto _out }

	goto _again
tr31:
	 m.cs = 0
//line plugins/parsers/influx/machine.go.rl:49

	m.err = ErrTimestampParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:56

	m.err = ErrParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++; goto _out }

	goto _again
tr50:
	 m.cs = 0
//line plugins/parsers/influx/machine.go.rl:42

	m.err = ErrTagParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:56

	m.err = ErrParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++; goto _out }

	goto _again
tr59:
	 m.cs = 0
//line plugins/parsers/influx/machine.go.rl:42

	m.err = ErrTagParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:35

	m.err = ErrFieldParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:56

	m.err = ErrParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++; goto _out }

	goto _again
tr99:
	 m.cs = 0
//line plugins/parsers/influx/machine.go.rl:35

	m.err = ErrFieldParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:49

	m.err = ErrTimestampParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:56

	m.err = ErrParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++; goto _out }

	goto _again
tr201:
	 m.cs = 0
//line plugins/parsers/influx/machine.go.rl:42

	m.err = ErrTagParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:35

	m.err = ErrFieldParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:49

	m.err = ErrTimestampParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:56

	m.err = ErrParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++; goto _out }

	goto _again
tr210:
	 m.cs = 0
//line plugins/parsers/influx/machine.go.rl:42

	m.err = ErrTagParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:49

	m.err = ErrTimestampParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:56

	m.err = ErrParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++; goto _out }

	goto _again
//line plugins/parsers/influx/machine.go:2761
st_case_0:
	st0:
		 m.cs = 0
		goto _out
tr14:
//line plugins/parsers/influx/machine.go.rl:84

	key = m.text()

	goto st5
	st5:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof5
		}
	st_case_5:
//line plugins/parsers/influx/machine.go:2777
		switch ( m.data)[( m.p)] {
		case 34:
			goto st6
		case 45:
			goto tr17
		case 46:
			goto tr18
		case 48:
			goto tr19
		case 70:
			goto tr21
		case 84:
			goto tr22
		case 102:
			goto tr23
		case 116:
			goto tr24
		}
		if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr20
		}
		goto tr5
	st6:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof6
		}
	st_case_6:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr26
		case 92:
			goto tr27
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto tr25
tr25:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st7
	st7:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof7
		}
	st_case_7:
//line plugins/parsers/influx/machine.go:2828
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st7
tr26:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st191
tr29:
//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st191
	st191:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof191
		}
	st_case_191:
//line plugins/parsers/influx/machine.go:2862
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 13:
			goto tr330
		case 32:
			goto st192
		case 44:
			goto st9
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st192
		}
		goto tr99
tr355:
//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

	goto st192
tr361:
//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

	goto st192
tr365:
//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

	goto st192
tr369:
//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

	goto st192
	st192:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof192
		}
	st_case_192:
//line plugins/parsers/influx/machine.go:2906
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 13:
			goto tr330
		case 32:
			goto st192
		case 45:
			goto tr332
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr333
			}
		case ( m.data)[( m.p)] >= 9:
			goto st192
		}
		goto tr31
tr330:
	 m.cs = 193
//line plugins/parsers/influx/machine.go.rl:22

	yield = true
	 m.cs = 188;
	{( m.p)++; goto _out }

	goto _again
tr335:
	 m.cs = 193
//line plugins/parsers/influx/machine.go.rl:108

	m.handler.SetTimestamp(m.text())

//line plugins/parsers/influx/machine.go.rl:22

	yield = true
	 m.cs = 188;
	{( m.p)++; goto _out }

	goto _again
tr356:
	 m.cs = 193
//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

//line plugins/parsers/influx/machine.go.rl:22

	yield = true
	 m.cs = 188;
	{( m.p)++; goto _out }

	goto _again
tr362:
	 m.cs = 193
//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

//line plugins/parsers/influx/machine.go.rl:22

	yield = true
	 m.cs = 188;
	{( m.p)++; goto _out }

	goto _again
tr366:
	 m.cs = 193
//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

//line plugins/parsers/influx/machine.go.rl:22

	yield = true
	 m.cs = 188;
	{( m.p)++; goto _out }

	goto _again
tr370:
	 m.cs = 193
//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

//line plugins/parsers/influx/machine.go.rl:22

	yield = true
	 m.cs = 188;
	{( m.p)++; goto _out }

	goto _again
	st193:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof193
		}
	st_case_193:
//line plugins/parsers/influx/machine.go:3005
		goto tr1
tr332:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st8
	st8:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof8
		}
	st_case_8:
//line plugins/parsers/influx/machine.go:3018
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st194
		}
		goto tr31
tr333:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st194
	st194:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof194
		}
	st_case_194:
//line plugins/parsers/influx/machine.go:3034
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 13:
			goto tr335
		case 32:
			goto tr334
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st196
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto tr31
tr334:
//line plugins/parsers/influx/machine.go.rl:108

	m.handler.SetTimestamp(m.text())

	goto st195
	st195:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof195
		}
	st_case_195:
//line plugins/parsers/influx/machine.go:3063
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 13:
			goto tr330
		case 32:
			goto st195
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st195
		}
		goto tr1
	st196:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof196
		}
	st_case_196:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 13:
			goto tr335
		case 32:
			goto tr334
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st197
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto tr31
	st197:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof197
		}
	st_case_197:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 13:
			goto tr335
		case 32:
			goto tr334
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st198
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto tr31
	st198:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof198
		}
	st_case_198:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 13:
			goto tr335
		case 32:
			goto tr334
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st199
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto tr31
	st199:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof199
		}
	st_case_199:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 13:
			goto tr335
		case 32:
			goto tr334
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st200
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto tr31
	st200:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof200
		}
	st_case_200:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 13:
			goto tr335
		case 32:
			goto tr334
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st201
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto tr31
	st201:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof201
		}
	st_case_201:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 13:
			goto tr335
		case 32:
			goto tr334
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st202
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto tr31
	st202:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof202
		}
	st_case_202:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 13:
			goto tr335
		case 32:
			goto tr334
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st203
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto tr31
	st203:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof203
		}
	st_case_203:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 13:
			goto tr335
		case 32:
			goto tr334
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st204
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto tr31
	st204:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof204
		}
	st_case_204:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 13:
			goto tr335
		case 32:
			goto tr334
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st205
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto tr31
	st205:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof205
		}
	st_case_205:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 13:
			goto tr335
		case 32:
			goto tr334
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st206
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto tr31
	st206:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof206
		}
	st_case_206:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 13:
			goto tr335
		case 32:
			goto tr334
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st207
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto tr31
	st207:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof207
		}
	st_case_207:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 13:
			goto tr335
		case 32:
			goto tr334
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st208
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto tr31
	st208:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof208
		}
	st_case_208:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 13:
			goto tr335
		case 32:
			goto tr334
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st209
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto tr31
	st209:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof209
		}
	st_case_209:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 13:
			goto tr335
		case 32:
			goto tr334
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st210
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto tr31
	st210:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof210
		}
	st_case_210:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 13:
			goto tr335
		case 32:
			goto tr334
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st211
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto tr31
	st211:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof211
		}
	st_case_211:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 13:
			goto tr335
		case 32:
			goto tr334
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st212
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto tr31
	st212:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof212
		}
	st_case_212:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 13:
			goto tr335
		case 32:
			goto tr334
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st213
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto tr31
	st213:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof213
		}
	st_case_213:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 13:
			goto tr335
		case 32:
			goto tr334
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr334
		}
		goto tr31
tr357:
//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

	goto st9
tr363:
//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

	goto st9
tr367:
//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

	goto st9
tr371:
//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

	goto st9
	st9:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof9
		}
	st_case_9:
//line plugins/parsers/influx/machine.go:3496
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr5
		case 44:
			goto tr5
		case 61:
			goto tr5
		case 92:
			goto tr12
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr5
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr5
		}
		goto tr9
tr12:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st10
	st10:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof10
		}
	st_case_10:
//line plugins/parsers/influx/machine.go:3527
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr5
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr5
		}
		goto st4
tr27:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st11
	st11:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof11
		}
	st_case_11:
//line plugins/parsers/influx/machine.go:3548
		switch ( m.data)[( m.p)] {
		case 34:
			goto st7
		case 92:
			goto st7
		}
		goto tr5
tr17:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st12
	st12:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof12
		}
	st_case_12:
//line plugins/parsers/influx/machine.go:3567
		if ( m.data)[( m.p)] == 48 {
			goto st214
		}
		if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st218
		}
		goto tr5
	st214:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof214
		}
	st_case_214:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 13:
			goto tr356
		case 32:
			goto tr355
		case 44:
			goto tr357
		case 46:
			goto st215
		case 69:
			goto st13
		case 101:
			goto st13
		case 105:
			goto st217
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr355
		}
		goto tr99
tr18:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st215
	st215:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof215
		}
	st_case_215:
//line plugins/parsers/influx/machine.go:3613
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 13:
			goto tr356
		case 32:
			goto tr355
		case 44:
			goto tr357
		case 69:
			goto st13
		case 101:
			goto st13
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st215
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr355
		}
		goto tr99
	st13:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof13
		}
	st_case_13:
		switch ( m.data)[( m.p)] {
		case 34:
			goto st14
		case 43:
			goto st14
		case 45:
			goto st14
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st216
		}
		goto tr5
	st14:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof14
		}
	st_case_14:
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st216
		}
		goto tr5
	st216:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof216
		}
	st_case_216:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 13:
			goto tr356
		case 32:
			goto tr355
		case 44:
			goto tr357
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st216
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr355
		}
		goto tr99
	st217:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof217
		}
	st_case_217:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr362
		case 13:
			goto tr362
		case 32:
			goto tr361
		case 44:
			goto tr363
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr361
		}
		goto tr99
	st218:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof218
		}
	st_case_218:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 13:
			goto tr356
		case 32:
			goto tr355
		case 44:
			goto tr357
		case 46:
			goto st215
		case 69:
			goto st13
		case 101:
			goto st13
		case 105:
			goto st217
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st218
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr355
		}
		goto tr99
tr19:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st219
	st219:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof219
		}
	st_case_219:
//line plugins/parsers/influx/machine.go:3749
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 13:
			goto tr356
		case 32:
			goto tr355
		case 44:
			goto tr357
		case 46:
			goto st215
		case 69:
			goto st13
		case 101:
			goto st13
		case 105:
			goto st217
		case 117:
			goto st220
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr355
		}
		goto tr99
	st220:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof220
		}
	st_case_220:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr366
		case 13:
			goto tr366
		case 32:
			goto tr365
		case 44:
			goto tr367
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr365
		}
		goto tr99
tr20:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st221
	st221:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof221
		}
	st_case_221:
//line plugins/parsers/influx/machine.go:3804
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 13:
			goto tr356
		case 32:
			goto tr355
		case 44:
			goto tr357
		case 46:
			goto st215
		case 69:
			goto st13
		case 101:
			goto st13
		case 105:
			goto st217
		case 117:
			goto st220
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st221
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr355
		}
		goto tr99
tr21:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st222
	st222:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof222
		}
	st_case_222:
//line plugins/parsers/influx/machine.go:3845
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 13:
			goto tr370
		case 32:
			goto tr369
		case 44:
			goto tr371
		case 65:
			goto st15
		case 97:
			goto st18
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr369
		}
		goto tr99
	st15:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof15
		}
	st_case_15:
		if ( m.data)[( m.p)] == 76 {
			goto st16
		}
		goto tr5
	st16:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof16
		}
	st_case_16:
		if ( m.data)[( m.p)] == 83 {
			goto st17
		}
		goto tr5
	st17:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof17
		}
	st_case_17:
		if ( m.data)[( m.p)] == 69 {
			goto st223
		}
		goto tr5
	st223:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof223
		}
	st_case_223:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 13:
			goto tr370
		case 32:
			goto tr369
		case 44:
			goto tr371
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr369
		}
		goto tr99
	st18:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof18
		}
	st_case_18:
		if ( m.data)[( m.p)] == 108 {
			goto st19
		}
		goto tr5
	st19:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof19
		}
	st_case_19:
		if ( m.data)[( m.p)] == 115 {
			goto st20
		}
		goto tr5
	st20:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof20
		}
	st_case_20:
		if ( m.data)[( m.p)] == 101 {
			goto st223
		}
		goto tr5
tr22:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st224
	st224:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof224
		}
	st_case_224:
//line plugins/parsers/influx/machine.go:3948
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 13:
			goto tr370
		case 32:
			goto tr369
		case 44:
			goto tr371
		case 82:
			goto st21
		case 114:
			goto st22
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr369
		}
		goto tr99
	st21:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof21
		}
	st_case_21:
		if ( m.data)[( m.p)] == 85 {
			goto st17
		}
		goto tr5
	st22:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof22
		}
	st_case_22:
		if ( m.data)[( m.p)] == 117 {
			goto st20
		}
		goto tr5
tr23:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st225
	st225:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof225
		}
	st_case_225:
//line plugins/parsers/influx/machine.go:3996
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 13:
			goto tr370
		case 32:
			goto tr369
		case 44:
			goto tr371
		case 97:
			goto st18
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr369
		}
		goto tr99
tr24:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st226
	st226:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof226
		}
	st_case_226:
//line plugins/parsers/influx/machine.go:4024
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 13:
			goto tr370
		case 32:
			goto tr369
		case 44:
			goto tr371
		case 114:
			goto st22
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr369
		}
		goto tr99
tr11:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st23
	st23:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof23
		}
	st_case_23:
//line plugins/parsers/influx/machine.go:4052
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 11:
			goto tr11
		case 13:
			goto tr5
		case 32:
			goto st3
		case 44:
			goto tr5
		case 61:
			goto tr14
		case 92:
			goto tr12
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st3
		}
		goto tr9
tr6:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

	goto st24
	st24:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof24
		}
	st_case_24:
//line plugins/parsers/influx/machine.go:4084
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 11:
			goto tr43
		case 13:
			goto tr5
		case 32:
			goto tr4
		case 44:
			goto tr7
		case 61:
			goto st2
		case 92:
			goto tr44
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr4
		}
		goto tr42
tr42:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st25
	st25:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof25
		}
	st_case_25:
//line plugins/parsers/influx/machine.go:4116
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 11:
			goto tr46
		case 13:
			goto tr5
		case 32:
			goto tr4
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr4
		}
		goto st25
tr46:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

	goto st26
tr43:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st26
	st26:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof26
		}
	st_case_26:
//line plugins/parsers/influx/machine.go:4158
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 11:
			goto tr43
		case 13:
			goto tr5
		case 32:
			goto tr4
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto tr44
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr4
		}
		goto tr42
tr7:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

	goto st27
tr61:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st27
	st27:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof27
		}
	st_case_27:
//line plugins/parsers/influx/machine.go:4196
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr50
		case 44:
			goto tr50
		case 61:
			goto tr50
		case 92:
			goto tr51
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr50
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr50
		}
		goto tr49
tr49:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st28
	st28:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof28
		}
	st_case_28:
//line plugins/parsers/influx/machine.go:4227
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr50
		case 44:
			goto tr50
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr50
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr50
		}
		goto st28
tr53:
//line plugins/parsers/influx/machine.go.rl:76

	key = m.text()

	goto st29
	st29:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof29
		}
	st_case_29:
//line plugins/parsers/influx/machine.go:4258
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr50
		case 44:
			goto tr50
		case 61:
			goto tr50
		case 92:
			goto tr56
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr50
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr50
		}
		goto tr55
tr55:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st30
	st30:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof30
		}
	st_case_30:
//line plugins/parsers/influx/machine.go:4289
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr59
		case 11:
			goto tr60
		case 13:
			goto tr59
		case 32:
			goto tr58
		case 44:
			goto tr61
		case 61:
			goto tr59
		case 92:
			goto st35
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr58
		}
		goto st30
tr60:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st31
	st31:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof31
		}
	st_case_31:
//line plugins/parsers/influx/machine.go:4321
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr59
		case 11:
			goto tr64
		case 13:
			goto tr59
		case 32:
			goto tr58
		case 44:
			goto tr61
		case 61:
			goto tr59
		case 92:
			goto tr65
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr58
		}
		goto tr63
tr63:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st32
	st32:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof32
		}
	st_case_32:
//line plugins/parsers/influx/machine.go:4353
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr59
		case 11:
			goto tr67
		case 13:
			goto tr59
		case 32:
			goto tr58
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr58
		}
		goto st32
tr67:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st33
tr64:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st33
	st33:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof33
		}
	st_case_33:
//line plugins/parsers/influx/machine.go:4395
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr59
		case 11:
			goto tr64
		case 13:
			goto tr59
		case 32:
			goto tr58
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto tr65
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr58
		}
		goto tr63
tr65:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st34
	st34:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof34
		}
	st_case_34:
//line plugins/parsers/influx/machine.go:4427
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr59
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr59
		}
		goto st32
tr56:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st35
	st35:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof35
		}
	st_case_35:
//line plugins/parsers/influx/machine.go:4448
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr50
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr50
		}
		goto st30
tr51:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st36
	st36:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof36
		}
	st_case_36:
//line plugins/parsers/influx/machine.go:4469
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr50
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr50
		}
		goto st28
tr47:
//line plugins/parsers/influx/machine.go.rl:84

	key = m.text()

	goto st37
	st37:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof37
		}
	st_case_37:
//line plugins/parsers/influx/machine.go:4490
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 11:
			goto tr6
		case 13:
			goto tr5
		case 32:
			goto tr4
		case 34:
			goto st38
		case 44:
			goto tr7
		case 45:
			goto tr70
		case 46:
			goto tr71
		case 48:
			goto tr72
		case 70:
			goto tr74
		case 84:
			goto tr75
		case 92:
			goto st129
		case 102:
			goto tr76
		case 116:
			goto tr77
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr73
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr4
		}
		goto st2
	st38:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof38
		}
	st_case_38:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr79
		case 11:
			goto tr80
		case 12:
			goto tr4
		case 32:
			goto tr79
		case 34:
			goto tr81
		case 44:
			goto tr82
		case 92:
			goto tr83
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto tr78
tr78:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st39
	st39:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof39
		}
	st_case_39:
//line plugins/parsers/influx/machine.go:4566
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr85
		case 11:
			goto tr86
		case 12:
			goto tr4
		case 32:
			goto tr85
		case 34:
			goto tr87
		case 44:
			goto tr88
		case 92:
			goto st164
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st39
tr85:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

	goto st40
tr79:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st40
tr229:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st40
	st40:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof40
		}
	st_case_40:
//line plugins/parsers/influx/machine.go:4614
		switch ( m.data)[( m.p)] {
		case 9:
			goto st40
		case 11:
			goto tr92
		case 12:
			goto st3
		case 32:
			goto st40
		case 34:
			goto tr93
		case 44:
			goto st7
		case 61:
			goto st7
		case 92:
			goto tr94
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto tr90
tr90:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st41
	st41:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof41
		}
	st_case_41:
//line plugins/parsers/influx/machine.go:4648
		switch ( m.data)[( m.p)] {
		case 9:
			goto st7
		case 10:
			goto tr5
		case 32:
			goto st7
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st41
tr93:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st227
tr96:
//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st227
tr112:
//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st227
	st227:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof227
		}
	st_case_227:
//line plugins/parsers/influx/machine.go:4700
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto st228
		case 13:
			goto tr330
		case 32:
			goto st192
		case 44:
			goto st9
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st192
		}
		goto st4
	st228:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof228
		}
	st_case_228:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto st228
		case 13:
			goto tr330
		case 32:
			goto st192
		case 44:
			goto tr99
		case 45:
			goto tr377
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr378
			}
		case ( m.data)[( m.p)] >= 9:
			goto st192
		}
		goto st4
tr377:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st42
	st42:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof42
		}
	st_case_42:
//line plugins/parsers/influx/machine.go:4764
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr99
		case 44:
			goto tr99
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] < 12:
			if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 10 {
				goto tr99
			}
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st229
			}
		default:
			goto tr99
		}
		goto st4
tr378:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st229
	st229:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof229
		}
	st_case_229:
//line plugins/parsers/influx/machine.go:4799
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr379
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr99
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st231
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st4
tr379:
//line plugins/parsers/influx/machine.go.rl:108

	m.handler.SetTimestamp(m.text())

	goto st230
	st230:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof230
		}
	st_case_230:
//line plugins/parsers/influx/machine.go:4836
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto st230
		case 13:
			goto tr330
		case 32:
			goto st195
		case 44:
			goto tr5
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st195
		}
		goto st4
	st231:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof231
		}
	st_case_231:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr379
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr99
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st232
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st4
	st232:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof232
		}
	st_case_232:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr379
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr99
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st233
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st4
	st233:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof233
		}
	st_case_233:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr379
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr99
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st234
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st4
	st234:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof234
		}
	st_case_234:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr379
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr99
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st235
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st4
	st235:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof235
		}
	st_case_235:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr379
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr99
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st236
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st4
	st236:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof236
		}
	st_case_236:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr379
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr99
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st237
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st4
	st237:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof237
		}
	st_case_237:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr379
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr99
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st238
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st4
	st238:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof238
		}
	st_case_238:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr379
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr99
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st239
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st4
	st239:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof239
		}
	st_case_239:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr379
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr99
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st240
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st4
	st240:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof240
		}
	st_case_240:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr379
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr99
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st241
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st4
	st241:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof241
		}
	st_case_241:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr379
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr99
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st242
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st4
	st242:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof242
		}
	st_case_242:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr379
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr99
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st243
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st4
	st243:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof243
		}
	st_case_243:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr379
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr99
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st244
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st4
	st244:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof244
		}
	st_case_244:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr379
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr99
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st245
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st4
	st245:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof245
		}
	st_case_245:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr379
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr99
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st246
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st4
	st246:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof246
		}
	st_case_246:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr379
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr99
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st247
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st4
	st247:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof247
		}
	st_case_247:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr379
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr99
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st248
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st4
	st248:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof248
		}
	st_case_248:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr379
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr99
		case 61:
			goto tr14
		case 92:
			goto st10
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr334
		}
		goto st4
tr97:
//line plugins/parsers/influx/machine.go.rl:84

	key = m.text()

	goto st43
	st43:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof43
		}
	st_case_43:
//line plugins/parsers/influx/machine.go:5403
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr101
		case 45:
			goto tr102
		case 46:
			goto tr103
		case 48:
			goto tr104
		case 70:
			goto tr106
		case 84:
			goto tr107
		case 92:
			goto st11
		case 102:
			goto tr108
		case 116:
			goto tr109
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr105
			}
		case ( m.data)[( m.p)] >= 12:
			goto tr5
		}
		goto st7
tr101:
//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st249
	st249:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof249
		}
	st_case_249:
//line plugins/parsers/influx/machine.go:5446
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 12:
			goto st192
		case 13:
			goto tr330
		case 32:
			goto tr399
		case 34:
			goto tr26
		case 44:
			goto tr400
		case 92:
			goto tr27
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr399
		}
		goto tr25
tr399:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st250
tr425:
//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

	goto st250
tr430:
//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

	goto st250
tr433:
//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

	goto st250
tr436:
//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

	goto st250
	st250:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof250
		}
	st_case_250:
//line plugins/parsers/influx/machine.go:5502
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 12:
			goto st192
		case 13:
			goto tr330
		case 32:
			goto st250
		case 34:
			goto tr29
		case 45:
			goto tr402
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr403
			}
		case ( m.data)[( m.p)] >= 9:
			goto st250
		}
		goto st7
tr402:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st44
	st44:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof44
		}
	st_case_44:
//line plugins/parsers/influx/machine.go:5539
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr99
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st251
			}
		case ( m.data)[( m.p)] >= 12:
			goto tr99
		}
		goto st7
tr403:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st251
	st251:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof251
		}
	st_case_251:
//line plugins/parsers/influx/machine.go:5568
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 12:
			goto tr334
		case 13:
			goto tr335
		case 32:
			goto tr404
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st253
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr404
		}
		goto st7
tr404:
//line plugins/parsers/influx/machine.go.rl:108

	m.handler.SetTimestamp(m.text())

	goto st252
	st252:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof252
		}
	st_case_252:
//line plugins/parsers/influx/machine.go:5603
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 12:
			goto st195
		case 13:
			goto tr330
		case 32:
			goto st252
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto st252
		}
		goto st7
	st253:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof253
		}
	st_case_253:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 12:
			goto tr334
		case 13:
			goto tr335
		case 32:
			goto tr404
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st254
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr404
		}
		goto st7
	st254:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof254
		}
	st_case_254:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 12:
			goto tr334
		case 13:
			goto tr335
		case 32:
			goto tr404
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st255
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr404
		}
		goto st7
	st255:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof255
		}
	st_case_255:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 12:
			goto tr334
		case 13:
			goto tr335
		case 32:
			goto tr404
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st256
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr404
		}
		goto st7
	st256:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof256
		}
	st_case_256:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 12:
			goto tr334
		case 13:
			goto tr335
		case 32:
			goto tr404
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st257
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr404
		}
		goto st7
	st257:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof257
		}
	st_case_257:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 12:
			goto tr334
		case 13:
			goto tr335
		case 32:
			goto tr404
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st258
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr404
		}
		goto st7
	st258:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof258
		}
	st_case_258:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 12:
			goto tr334
		case 13:
			goto tr335
		case 32:
			goto tr404
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st259
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr404
		}
		goto st7
	st259:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof259
		}
	st_case_259:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 12:
			goto tr334
		case 13:
			goto tr335
		case 32:
			goto tr404
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st260
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr404
		}
		goto st7
	st260:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof260
		}
	st_case_260:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 12:
			goto tr334
		case 13:
			goto tr335
		case 32:
			goto tr404
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st261
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr404
		}
		goto st7
	st261:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof261
		}
	st_case_261:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 12:
			goto tr334
		case 13:
			goto tr335
		case 32:
			goto tr404
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st262
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr404
		}
		goto st7
	st262:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof262
		}
	st_case_262:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 12:
			goto tr334
		case 13:
			goto tr335
		case 32:
			goto tr404
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st263
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr404
		}
		goto st7
	st263:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof263
		}
	st_case_263:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 12:
			goto tr334
		case 13:
			goto tr335
		case 32:
			goto tr404
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st264
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr404
		}
		goto st7
	st264:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof264
		}
	st_case_264:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 12:
			goto tr334
		case 13:
			goto tr335
		case 32:
			goto tr404
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st265
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr404
		}
		goto st7
	st265:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof265
		}
	st_case_265:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 12:
			goto tr334
		case 13:
			goto tr335
		case 32:
			goto tr404
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st266
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr404
		}
		goto st7
	st266:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof266
		}
	st_case_266:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 12:
			goto tr334
		case 13:
			goto tr335
		case 32:
			goto tr404
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st267
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr404
		}
		goto st7
	st267:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof267
		}
	st_case_267:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 12:
			goto tr334
		case 13:
			goto tr335
		case 32:
			goto tr404
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st268
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr404
		}
		goto st7
	st268:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof268
		}
	st_case_268:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 12:
			goto tr334
		case 13:
			goto tr335
		case 32:
			goto tr404
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st269
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr404
		}
		goto st7
	st269:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof269
		}
	st_case_269:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 12:
			goto tr334
		case 13:
			goto tr335
		case 32:
			goto tr404
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st270
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr404
		}
		goto st7
	st270:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof270
		}
	st_case_270:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 12:
			goto tr334
		case 13:
			goto tr335
		case 32:
			goto tr404
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr404
		}
		goto st7
tr400:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st45
tr442:
//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

	goto st45
tr446:
//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

	goto st45
tr448:
//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

	goto st45
tr450:
//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

	goto st45
	st45:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof45
		}
	st_case_45:
//line plugins/parsers/influx/machine.go:6156
		switch ( m.data)[( m.p)] {
		case 9:
			goto st7
		case 10:
			goto tr5
		case 32:
			goto st7
		case 34:
			goto tr112
		case 44:
			goto st7
		case 61:
			goto st7
		case 92:
			goto tr113
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto tr111
tr111:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st46
	st46:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof46
		}
	st_case_46:
//line plugins/parsers/influx/machine.go:6188
		switch ( m.data)[( m.p)] {
		case 9:
			goto st7
		case 10:
			goto tr5
		case 32:
			goto st7
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr115
		case 92:
			goto st74
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st46
tr115:
//line plugins/parsers/influx/machine.go.rl:84

	key = m.text()

	goto st47
	st47:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof47
		}
	st_case_47:
//line plugins/parsers/influx/machine.go:6220
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr117
		case 45:
			goto tr102
		case 46:
			goto tr103
		case 48:
			goto tr104
		case 70:
			goto tr106
		case 84:
			goto tr107
		case 92:
			goto st11
		case 102:
			goto tr108
		case 116:
			goto tr109
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr105
			}
		case ( m.data)[( m.p)] >= 12:
			goto tr5
		}
		goto st7
tr117:
//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st271
	st271:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof271
		}
	st_case_271:
//line plugins/parsers/influx/machine.go:6263
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 12:
			goto st192
		case 13:
			goto tr330
		case 32:
			goto tr399
		case 34:
			goto tr26
		case 44:
			goto tr424
		case 92:
			goto tr27
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr399
		}
		goto tr25
tr424:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st48
tr426:
//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

	goto st48
tr431:
//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

	goto st48
tr434:
//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

	goto st48
tr437:
//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

	goto st48
	st48:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof48
		}
	st_case_48:
//line plugins/parsers/influx/machine.go:6319
		switch ( m.data)[( m.p)] {
		case 9:
			goto st7
		case 10:
			goto tr5
		case 32:
			goto st7
		case 34:
			goto tr93
		case 44:
			goto st7
		case 61:
			goto st7
		case 92:
			goto tr119
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto tr118
tr118:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st49
	st49:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof49
		}
	st_case_49:
//line plugins/parsers/influx/machine.go:6351
		switch ( m.data)[( m.p)] {
		case 9:
			goto st7
		case 10:
			goto tr5
		case 32:
			goto st7
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr121
		case 92:
			goto st62
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st49
tr121:
//line plugins/parsers/influx/machine.go.rl:84

	key = m.text()

	goto st50
	st50:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof50
		}
	st_case_50:
//line plugins/parsers/influx/machine.go:6383
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr117
		case 45:
			goto tr123
		case 46:
			goto tr124
		case 48:
			goto tr125
		case 70:
			goto tr127
		case 84:
			goto tr128
		case 92:
			goto st11
		case 102:
			goto tr129
		case 116:
			goto tr130
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr126
			}
		case ( m.data)[( m.p)] >= 12:
			goto tr5
		}
		goto st7
tr123:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st51
	st51:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof51
		}
	st_case_51:
//line plugins/parsers/influx/machine.go:6426
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 48:
			goto st272
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st277
			}
		case ( m.data)[( m.p)] >= 12:
			goto tr5
		}
		goto st7
	st272:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof272
		}
	st_case_272:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 12:
			goto tr355
		case 13:
			goto tr356
		case 32:
			goto tr425
		case 34:
			goto tr29
		case 44:
			goto tr426
		case 46:
			goto st273
		case 69:
			goto st52
		case 92:
			goto st11
		case 101:
			goto st52
		case 105:
			goto st276
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr425
		}
		goto st7
tr124:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st273
	st273:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof273
		}
	st_case_273:
//line plugins/parsers/influx/machine.go:6490
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 12:
			goto tr355
		case 13:
			goto tr356
		case 32:
			goto tr425
		case 34:
			goto tr29
		case 44:
			goto tr426
		case 69:
			goto st52
		case 92:
			goto st11
		case 101:
			goto st52
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st273
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr425
		}
		goto st7
	st52:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof52
		}
	st_case_52:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr133
		case 43:
			goto st53
		case 45:
			goto st53
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st275
			}
		case ( m.data)[( m.p)] >= 12:
			goto tr5
		}
		goto st7
tr133:
//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st274
	st274:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof274
		}
	st_case_274:
//line plugins/parsers/influx/machine.go:6557
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 13:
			goto tr330
		case 32:
			goto st192
		case 44:
			goto st9
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st216
			}
		case ( m.data)[( m.p)] >= 9:
			goto st192
		}
		goto tr99
	st53:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof53
		}
	st_case_53:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st275
			}
		case ( m.data)[( m.p)] >= 12:
			goto tr5
		}
		goto st7
	st275:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof275
		}
	st_case_275:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 12:
			goto tr355
		case 13:
			goto tr356
		case 32:
			goto tr425
		case 34:
			goto tr29
		case 44:
			goto tr426
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st275
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr425
		}
		goto st7
	st276:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof276
		}
	st_case_276:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr362
		case 12:
			goto tr361
		case 13:
			goto tr362
		case 32:
			goto tr430
		case 34:
			goto tr29
		case 44:
			goto tr431
		case 92:
			goto st11
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr430
		}
		goto st7
	st277:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof277
		}
	st_case_277:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 12:
			goto tr355
		case 13:
			goto tr356
		case 32:
			goto tr425
		case 34:
			goto tr29
		case 44:
			goto tr426
		case 46:
			goto st273
		case 69:
			goto st52
		case 92:
			goto st11
		case 101:
			goto st52
		case 105:
			goto st276
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st277
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr425
		}
		goto st7
tr125:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st278
	st278:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof278
		}
	st_case_278:
//line plugins/parsers/influx/machine.go:6703
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 12:
			goto tr355
		case 13:
			goto tr356
		case 32:
			goto tr425
		case 34:
			goto tr29
		case 44:
			goto tr426
		case 46:
			goto st273
		case 69:
			goto st52
		case 92:
			goto st11
		case 101:
			goto st52
		case 105:
			goto st276
		case 117:
			goto st279
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr425
		}
		goto st7
	st279:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof279
		}
	st_case_279:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr366
		case 12:
			goto tr365
		case 13:
			goto tr366
		case 32:
			goto tr433
		case 34:
			goto tr29
		case 44:
			goto tr434
		case 92:
			goto st11
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr433
		}
		goto st7
tr126:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st280
	st280:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof280
		}
	st_case_280:
//line plugins/parsers/influx/machine.go:6770
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 12:
			goto tr355
		case 13:
			goto tr356
		case 32:
			goto tr425
		case 34:
			goto tr29
		case 44:
			goto tr426
		case 46:
			goto st273
		case 69:
			goto st52
		case 92:
			goto st11
		case 101:
			goto st52
		case 105:
			goto st276
		case 117:
			goto st279
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st280
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr425
		}
		goto st7
tr127:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st281
	st281:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof281
		}
	st_case_281:
//line plugins/parsers/influx/machine.go:6817
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 12:
			goto tr369
		case 13:
			goto tr370
		case 32:
			goto tr436
		case 34:
			goto tr29
		case 44:
			goto tr437
		case 65:
			goto st54
		case 92:
			goto st11
		case 97:
			goto st57
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr436
		}
		goto st7
	st54:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof54
		}
	st_case_54:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 76:
			goto st55
		case 92:
			goto st11
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st7
	st55:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof55
		}
	st_case_55:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 83:
			goto st56
		case 92:
			goto st11
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st7
	st56:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof56
		}
	st_case_56:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 69:
			goto st282
		case 92:
			goto st11
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st7
	st282:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof282
		}
	st_case_282:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 12:
			goto tr369
		case 13:
			goto tr370
		case 32:
			goto tr436
		case 34:
			goto tr29
		case 44:
			goto tr437
		case 92:
			goto st11
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr436
		}
		goto st7
	st57:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof57
		}
	st_case_57:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 92:
			goto st11
		case 108:
			goto st58
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st7
	st58:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof58
		}
	st_case_58:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 92:
			goto st11
		case 115:
			goto st59
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st7
	st59:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof59
		}
	st_case_59:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 92:
			goto st11
		case 101:
			goto st282
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st7
tr128:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st283
	st283:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof283
		}
	st_case_283:
//line plugins/parsers/influx/machine.go:6992
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 12:
			goto tr369
		case 13:
			goto tr370
		case 32:
			goto tr436
		case 34:
			goto tr29
		case 44:
			goto tr437
		case 82:
			goto st60
		case 92:
			goto st11
		case 114:
			goto st61
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr436
		}
		goto st7
	st60:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof60
		}
	st_case_60:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 85:
			goto st56
		case 92:
			goto st11
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st7
	st61:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof61
		}
	st_case_61:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 92:
			goto st11
		case 117:
			goto st59
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st7
tr129:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st284
	st284:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof284
		}
	st_case_284:
//line plugins/parsers/influx/machine.go:7066
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 12:
			goto tr369
		case 13:
			goto tr370
		case 32:
			goto tr436
		case 34:
			goto tr29
		case 44:
			goto tr437
		case 92:
			goto st11
		case 97:
			goto st57
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr436
		}
		goto st7
tr130:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st285
	st285:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof285
		}
	st_case_285:
//line plugins/parsers/influx/machine.go:7100
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 12:
			goto tr369
		case 13:
			goto tr370
		case 32:
			goto tr436
		case 34:
			goto tr29
		case 44:
			goto tr437
		case 92:
			goto st11
		case 114:
			goto st61
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr436
		}
		goto st7
tr119:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st62
	st62:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof62
		}
	st_case_62:
//line plugins/parsers/influx/machine.go:7134
		switch ( m.data)[( m.p)] {
		case 34:
			goto st49
		case 92:
			goto st49
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr5
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr5
		}
		goto st4
tr102:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st63
	st63:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof63
		}
	st_case_63:
//line plugins/parsers/influx/machine.go:7161
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 48:
			goto st286
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st290
			}
		case ( m.data)[( m.p)] >= 12:
			goto tr5
		}
		goto st7
	st286:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof286
		}
	st_case_286:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 12:
			goto tr355
		case 13:
			goto tr356
		case 32:
			goto tr425
		case 34:
			goto tr29
		case 44:
			goto tr442
		case 46:
			goto st287
		case 69:
			goto st64
		case 92:
			goto st11
		case 101:
			goto st64
		case 105:
			goto st289
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr425
		}
		goto st7
tr103:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st287
	st287:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof287
		}
	st_case_287:
//line plugins/parsers/influx/machine.go:7225
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 12:
			goto tr355
		case 13:
			goto tr356
		case 32:
			goto tr425
		case 34:
			goto tr29
		case 44:
			goto tr442
		case 69:
			goto st64
		case 92:
			goto st11
		case 101:
			goto st64
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st287
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr425
		}
		goto st7
	st64:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof64
		}
	st_case_64:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr133
		case 43:
			goto st65
		case 45:
			goto st65
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st288
			}
		case ( m.data)[( m.p)] >= 12:
			goto tr5
		}
		goto st7
	st65:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof65
		}
	st_case_65:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st288
			}
		case ( m.data)[( m.p)] >= 12:
			goto tr5
		}
		goto st7
	st288:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof288
		}
	st_case_288:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 12:
			goto tr355
		case 13:
			goto tr356
		case 32:
			goto tr425
		case 34:
			goto tr29
		case 44:
			goto tr442
		case 92:
			goto st11
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st288
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr425
		}
		goto st7
	st289:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof289
		}
	st_case_289:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr362
		case 12:
			goto tr361
		case 13:
			goto tr362
		case 32:
			goto tr430
		case 34:
			goto tr29
		case 44:
			goto tr446
		case 92:
			goto st11
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr430
		}
		goto st7
	st290:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof290
		}
	st_case_290:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 12:
			goto tr355
		case 13:
			goto tr356
		case 32:
			goto tr425
		case 34:
			goto tr29
		case 44:
			goto tr442
		case 46:
			goto st287
		case 69:
			goto st64
		case 92:
			goto st11
		case 101:
			goto st64
		case 105:
			goto st289
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st290
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr425
		}
		goto st7
tr104:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st291
	st291:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof291
		}
	st_case_291:
//line plugins/parsers/influx/machine.go:7407
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 12:
			goto tr355
		case 13:
			goto tr356
		case 32:
			goto tr425
		case 34:
			goto tr29
		case 44:
			goto tr442
		case 46:
			goto st287
		case 69:
			goto st64
		case 92:
			goto st11
		case 101:
			goto st64
		case 105:
			goto st289
		case 117:
			goto st292
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr425
		}
		goto st7
	st292:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof292
		}
	st_case_292:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr366
		case 12:
			goto tr365
		case 13:
			goto tr366
		case 32:
			goto tr433
		case 34:
			goto tr29
		case 44:
			goto tr448
		case 92:
			goto st11
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr433
		}
		goto st7
tr105:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st293
	st293:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof293
		}
	st_case_293:
//line plugins/parsers/influx/machine.go:7474
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 12:
			goto tr355
		case 13:
			goto tr356
		case 32:
			goto tr425
		case 34:
			goto tr29
		case 44:
			goto tr442
		case 46:
			goto st287
		case 69:
			goto st64
		case 92:
			goto st11
		case 101:
			goto st64
		case 105:
			goto st289
		case 117:
			goto st292
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st293
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr425
		}
		goto st7
tr106:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st294
	st294:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof294
		}
	st_case_294:
//line plugins/parsers/influx/machine.go:7521
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 12:
			goto tr369
		case 13:
			goto tr370
		case 32:
			goto tr436
		case 34:
			goto tr29
		case 44:
			goto tr450
		case 65:
			goto st66
		case 92:
			goto st11
		case 97:
			goto st69
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr436
		}
		goto st7
	st66:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof66
		}
	st_case_66:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 76:
			goto st67
		case 92:
			goto st11
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st7
	st67:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof67
		}
	st_case_67:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 83:
			goto st68
		case 92:
			goto st11
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st7
	st68:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof68
		}
	st_case_68:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 69:
			goto st295
		case 92:
			goto st11
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st7
	st295:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof295
		}
	st_case_295:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 12:
			goto tr369
		case 13:
			goto tr370
		case 32:
			goto tr436
		case 34:
			goto tr29
		case 44:
			goto tr450
		case 92:
			goto st11
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr436
		}
		goto st7
	st69:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof69
		}
	st_case_69:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 92:
			goto st11
		case 108:
			goto st70
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st7
	st70:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof70
		}
	st_case_70:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 92:
			goto st11
		case 115:
			goto st71
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st7
	st71:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof71
		}
	st_case_71:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 92:
			goto st11
		case 101:
			goto st295
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st7
tr107:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st296
	st296:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof296
		}
	st_case_296:
//line plugins/parsers/influx/machine.go:7696
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 12:
			goto tr369
		case 13:
			goto tr370
		case 32:
			goto tr436
		case 34:
			goto tr29
		case 44:
			goto tr450
		case 82:
			goto st72
		case 92:
			goto st11
		case 114:
			goto st73
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr436
		}
		goto st7
	st72:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof72
		}
	st_case_72:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 85:
			goto st68
		case 92:
			goto st11
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st7
	st73:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof73
		}
	st_case_73:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr29
		case 92:
			goto st11
		case 117:
			goto st71
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st7
tr108:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st297
	st297:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof297
		}
	st_case_297:
//line plugins/parsers/influx/machine.go:7770
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 12:
			goto tr369
		case 13:
			goto tr370
		case 32:
			goto tr436
		case 34:
			goto tr29
		case 44:
			goto tr450
		case 92:
			goto st11
		case 97:
			goto st69
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr436
		}
		goto st7
tr109:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st298
	st298:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof298
		}
	st_case_298:
//line plugins/parsers/influx/machine.go:7804
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 12:
			goto tr369
		case 13:
			goto tr370
		case 32:
			goto tr436
		case 34:
			goto tr29
		case 44:
			goto tr450
		case 92:
			goto st11
		case 114:
			goto st73
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr436
		}
		goto st7
tr113:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st74
	st74:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof74
		}
	st_case_74:
//line plugins/parsers/influx/machine.go:7838
		switch ( m.data)[( m.p)] {
		case 34:
			goto st46
		case 92:
			goto st46
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr5
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr5
		}
		goto st4
tr94:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st75
	st75:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof75
		}
	st_case_75:
//line plugins/parsers/influx/machine.go:7865
		switch ( m.data)[( m.p)] {
		case 34:
			goto st41
		case 92:
			goto st41
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr5
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr5
		}
		goto st4
tr92:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st76
	st76:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof76
		}
	st_case_76:
//line plugins/parsers/influx/machine.go:7892
		switch ( m.data)[( m.p)] {
		case 9:
			goto st40
		case 11:
			goto tr92
		case 12:
			goto st3
		case 32:
			goto st40
		case 34:
			goto tr93
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto tr94
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto tr90
tr86:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

	goto st77
tr80:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st77
	st77:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof77
		}
	st_case_77:
//line plugins/parsers/influx/machine.go:7936
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr85
		case 11:
			goto tr151
		case 12:
			goto tr4
		case 32:
			goto tr85
		case 34:
			goto tr152
		case 44:
			goto tr88
		case 61:
			goto st39
		case 92:
			goto tr153
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto tr150
tr150:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st78
	st78:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof78
		}
	st_case_78:
//line plugins/parsers/influx/machine.go:7970
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr85
		case 11:
			goto tr155
		case 12:
			goto tr4
		case 32:
			goto tr85
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st78
tr155:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

	goto st79
tr151:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st79
	st79:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof79
		}
	st_case_79:
//line plugins/parsers/influx/machine.go:8014
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr85
		case 11:
			goto tr151
		case 12:
			goto tr4
		case 32:
			goto tr85
		case 34:
			goto tr152
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto tr153
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto tr150
tr152:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st299
tr156:
//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st299
	st299:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof299
		}
	st_case_299:
//line plugins/parsers/influx/machine.go:8058
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto tr456
		case 13:
			goto tr330
		case 32:
			goto tr455
		case 44:
			goto tr457
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr455
		}
		goto st25
tr455:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

	goto st300
tr487:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st300
tr539:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

	goto st300
tr545:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

	goto st300
tr549:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

	goto st300
tr553:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

	goto st300
tr764:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

	goto st300
tr773:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

	goto st300
tr778:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

	goto st300
tr783:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

	goto st300
	st300:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof300
		}
	st_case_300:
//line plugins/parsers/influx/machine.go:8176
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto tr459
		case 13:
			goto tr330
		case 32:
			goto st300
		case 44:
			goto tr99
		case 45:
			goto tr377
		case 61:
			goto tr99
		case 92:
			goto tr12
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr378
			}
		case ( m.data)[( m.p)] >= 9:
			goto st300
		}
		goto tr9
tr459:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st301
	st301:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof301
		}
	st_case_301:
//line plugins/parsers/influx/machine.go:8215
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto tr459
		case 13:
			goto tr330
		case 32:
			goto st300
		case 44:
			goto tr99
		case 45:
			goto tr377
		case 61:
			goto tr14
		case 92:
			goto tr12
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr378
			}
		case ( m.data)[( m.p)] >= 9:
			goto st300
		}
		goto tr9
tr456:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

	goto st302
tr460:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st302
	st302:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof302
		}
	st_case_302:
//line plugins/parsers/influx/machine.go:8264
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto tr460
		case 13:
			goto tr330
		case 32:
			goto tr455
		case 44:
			goto tr7
		case 45:
			goto tr461
		case 61:
			goto tr47
		case 92:
			goto tr44
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr462
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr455
		}
		goto tr42
tr461:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st80
	st80:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof80
		}
	st_case_80:
//line plugins/parsers/influx/machine.go:8303
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr99
		case 11:
			goto tr46
		case 13:
			goto tr99
		case 32:
			goto tr4
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st303
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr4
		}
		goto st25
tr462:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st303
	st303:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof303
		}
	st_case_303:
//line plugins/parsers/influx/machine.go:8340
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr464
		case 13:
			goto tr335
		case 32:
			goto tr463
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st307
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr463
		}
		goto st25
tr468:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

	goto st304
tr496:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st304
tr463:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:108

	m.handler.SetTimestamp(m.text())

	goto st304
tr493:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:108

	m.handler.SetTimestamp(m.text())

	goto st304
	st304:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof304
		}
	st_case_304:
//line plugins/parsers/influx/machine.go:8403
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto tr467
		case 13:
			goto tr330
		case 32:
			goto st304
		case 44:
			goto tr5
		case 61:
			goto tr5
		case 92:
			goto tr12
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st304
		}
		goto tr9
tr467:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st305
	st305:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof305
		}
	st_case_305:
//line plugins/parsers/influx/machine.go:8435
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto tr467
		case 13:
			goto tr330
		case 32:
			goto st304
		case 44:
			goto tr5
		case 61:
			goto tr14
		case 92:
			goto tr12
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st304
		}
		goto tr9
tr469:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st306
tr464:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:108

	m.handler.SetTimestamp(m.text())

	goto st306
	st306:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof306
		}
	st_case_306:
//line plugins/parsers/influx/machine.go:8481
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto tr469
		case 13:
			goto tr330
		case 32:
			goto tr468
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto tr44
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr468
		}
		goto tr42
tr44:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st81
	st81:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof81
		}
	st_case_81:
//line plugins/parsers/influx/machine.go:8513
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr5
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr5
		}
		goto st25
	st307:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof307
		}
	st_case_307:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr464
		case 13:
			goto tr335
		case 32:
			goto tr463
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st308
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr463
		}
		goto st25
	st308:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof308
		}
	st_case_308:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr464
		case 13:
			goto tr335
		case 32:
			goto tr463
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st309
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr463
		}
		goto st25
	st309:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof309
		}
	st_case_309:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr464
		case 13:
			goto tr335
		case 32:
			goto tr463
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st310
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr463
		}
		goto st25
	st310:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof310
		}
	st_case_310:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr464
		case 13:
			goto tr335
		case 32:
			goto tr463
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st311
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr463
		}
		goto st25
	st311:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof311
		}
	st_case_311:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr464
		case 13:
			goto tr335
		case 32:
			goto tr463
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st312
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr463
		}
		goto st25
	st312:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof312
		}
	st_case_312:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr464
		case 13:
			goto tr335
		case 32:
			goto tr463
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st313
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr463
		}
		goto st25
	st313:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof313
		}
	st_case_313:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr464
		case 13:
			goto tr335
		case 32:
			goto tr463
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st314
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr463
		}
		goto st25
	st314:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof314
		}
	st_case_314:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr464
		case 13:
			goto tr335
		case 32:
			goto tr463
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st315
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr463
		}
		goto st25
	st315:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof315
		}
	st_case_315:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr464
		case 13:
			goto tr335
		case 32:
			goto tr463
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st316
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr463
		}
		goto st25
	st316:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof316
		}
	st_case_316:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr464
		case 13:
			goto tr335
		case 32:
			goto tr463
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st317
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr463
		}
		goto st25
	st317:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof317
		}
	st_case_317:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr464
		case 13:
			goto tr335
		case 32:
			goto tr463
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st318
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr463
		}
		goto st25
	st318:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof318
		}
	st_case_318:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr464
		case 13:
			goto tr335
		case 32:
			goto tr463
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st319
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr463
		}
		goto st25
	st319:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof319
		}
	st_case_319:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr464
		case 13:
			goto tr335
		case 32:
			goto tr463
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st320
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr463
		}
		goto st25
	st320:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof320
		}
	st_case_320:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr464
		case 13:
			goto tr335
		case 32:
			goto tr463
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st321
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr463
		}
		goto st25
	st321:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof321
		}
	st_case_321:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr464
		case 13:
			goto tr335
		case 32:
			goto tr463
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st322
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr463
		}
		goto st25
	st322:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof322
		}
	st_case_322:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr464
		case 13:
			goto tr335
		case 32:
			goto tr463
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st323
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr463
		}
		goto st25
	st323:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof323
		}
	st_case_323:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr464
		case 13:
			goto tr335
		case 32:
			goto tr463
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st324
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr463
		}
		goto st25
	st324:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof324
		}
	st_case_324:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr464
		case 13:
			goto tr335
		case 32:
			goto tr463
		case 44:
			goto tr7
		case 61:
			goto tr47
		case 92:
			goto st81
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr463
		}
		goto st25
tr457:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

	goto st82
tr489:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st82
tr541:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

	goto st82
tr547:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

	goto st82
tr551:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

	goto st82
tr555:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

	goto st82
tr770:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

	goto st82
tr793:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

	goto st82
tr796:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

	goto st82
tr799:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

	goto st82
	st82:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof82
		}
	st_case_82:
//line plugins/parsers/influx/machine.go:9155
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr59
		case 44:
			goto tr59
		case 61:
			goto tr59
		case 92:
			goto tr161
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr59
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr59
		}
		goto tr160
tr160:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st83
	st83:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof83
		}
	st_case_83:
//line plugins/parsers/influx/machine.go:9186
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr59
		case 44:
			goto tr59
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr59
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr59
		}
		goto st83
tr163:
//line plugins/parsers/influx/machine.go.rl:76

	key = m.text()

//line plugins/parsers/influx/machine.go.rl:84

	key = m.text()

	goto st84
	st84:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof84
		}
	st_case_84:
//line plugins/parsers/influx/machine.go:9221
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr59
		case 34:
			goto tr165
		case 44:
			goto tr59
		case 45:
			goto tr166
		case 46:
			goto tr167
		case 48:
			goto tr168
		case 61:
			goto tr59
		case 70:
			goto tr170
		case 84:
			goto tr171
		case 92:
			goto tr56
		case 102:
			goto tr172
		case 116:
			goto tr173
		}
		switch {
		case ( m.data)[( m.p)] < 12:
			if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 10 {
				goto tr59
			}
		case ( m.data)[( m.p)] > 13:
			if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr169
			}
		default:
			goto tr59
		}
		goto tr55
tr165:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st85
	st85:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof85
		}
	st_case_85:
//line plugins/parsers/influx/machine.go:9272
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr175
		case 11:
			goto tr176
		case 12:
			goto tr58
		case 32:
			goto tr175
		case 34:
			goto tr177
		case 44:
			goto tr178
		case 61:
			goto tr25
		case 92:
			goto tr179
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto tr174
tr174:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st86
	st86:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof86
		}
	st_case_86:
//line plugins/parsers/influx/machine.go:9306
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr181
		case 11:
			goto tr182
		case 12:
			goto tr58
		case 32:
			goto tr181
		case 34:
			goto tr183
		case 44:
			goto tr184
		case 61:
			goto st7
		case 92:
			goto st100
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st86
tr181:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st87
tr175:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st87
	st87:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof87
		}
	st_case_87:
//line plugins/parsers/influx/machine.go:9350
		switch ( m.data)[( m.p)] {
		case 9:
			goto st87
		case 11:
			goto tr188
		case 12:
			goto st3
		case 32:
			goto st87
		case 34:
			goto tr93
		case 44:
			goto st7
		case 61:
			goto st7
		case 92:
			goto tr189
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto tr186
tr186:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st88
	st88:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof88
		}
	st_case_88:
//line plugins/parsers/influx/machine.go:9384
		switch ( m.data)[( m.p)] {
		case 9:
			goto st7
		case 10:
			goto tr5
		case 32:
			goto st7
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st88
tr191:
//line plugins/parsers/influx/machine.go.rl:84

	key = m.text()

	goto st89
	st89:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof89
		}
	st_case_89:
//line plugins/parsers/influx/machine.go:9416
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 34:
			goto tr101
		case 45:
			goto tr123
		case 46:
			goto tr124
		case 48:
			goto tr125
		case 70:
			goto tr127
		case 84:
			goto tr128
		case 92:
			goto st11
		case 102:
			goto tr129
		case 116:
			goto tr130
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr126
			}
		case ( m.data)[( m.p)] >= 12:
			goto tr5
		}
		goto st7
tr189:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st90
	st90:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof90
		}
	st_case_90:
//line plugins/parsers/influx/machine.go:9459
		switch ( m.data)[( m.p)] {
		case 34:
			goto st88
		case 92:
			goto st88
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr5
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr5
		}
		goto st4
tr188:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st91
	st91:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof91
		}
	st_case_91:
//line plugins/parsers/influx/machine.go:9486
		switch ( m.data)[( m.p)] {
		case 9:
			goto st87
		case 11:
			goto tr188
		case 12:
			goto st3
		case 32:
			goto st87
		case 34:
			goto tr93
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto tr189
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto tr186
tr182:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st92
tr176:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st92
	st92:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof92
		}
	st_case_92:
//line plugins/parsers/influx/machine.go:9530
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr181
		case 11:
			goto tr194
		case 12:
			goto tr58
		case 32:
			goto tr181
		case 34:
			goto tr195
		case 44:
			goto tr184
		case 61:
			goto st7
		case 92:
			goto tr196
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto tr193
tr193:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st93
	st93:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof93
		}
	st_case_93:
//line plugins/parsers/influx/machine.go:9564
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr181
		case 11:
			goto tr198
		case 12:
			goto tr58
		case 32:
			goto tr181
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st93
tr198:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st94
tr194:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st94
	st94:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof94
		}
	st_case_94:
//line plugins/parsers/influx/machine.go:9608
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr181
		case 11:
			goto tr194
		case 12:
			goto tr58
		case 32:
			goto tr181
		case 34:
			goto tr195
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto tr196
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto tr193
tr195:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st325
tr199:
//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st325
	st325:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof325
		}
	st_case_325:
//line plugins/parsers/influx/machine.go:9652
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto tr488
		case 13:
			goto tr330
		case 32:
			goto tr487
		case 44:
			goto tr489
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr487
		}
		goto st32
tr488:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st326
tr490:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st326
	st326:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof326
		}
	st_case_326:
//line plugins/parsers/influx/machine.go:9694
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto tr490
		case 13:
			goto tr330
		case 32:
			goto tr487
		case 44:
			goto tr61
		case 45:
			goto tr491
		case 61:
			goto tr14
		case 92:
			goto tr65
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr492
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr487
		}
		goto tr63
tr491:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st95
	st95:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof95
		}
	st_case_95:
//line plugins/parsers/influx/machine.go:9733
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr201
		case 11:
			goto tr67
		case 13:
			goto tr201
		case 32:
			goto tr58
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st327
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr58
		}
		goto st32
tr492:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st327
	st327:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof327
		}
	st_case_327:
//line plugins/parsers/influx/machine.go:9770
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr494
		case 13:
			goto tr335
		case 32:
			goto tr493
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st329
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr493
		}
		goto st32
tr497:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st328
tr494:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:108

	m.handler.SetTimestamp(m.text())

	goto st328
	st328:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof328
		}
	st_case_328:
//line plugins/parsers/influx/machine.go:9821
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto tr497
		case 13:
			goto tr330
		case 32:
			goto tr496
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto tr65
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr496
		}
		goto tr63
	st329:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof329
		}
	st_case_329:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr494
		case 13:
			goto tr335
		case 32:
			goto tr493
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st330
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr493
		}
		goto st32
	st330:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof330
		}
	st_case_330:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr494
		case 13:
			goto tr335
		case 32:
			goto tr493
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st331
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr493
		}
		goto st32
	st331:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof331
		}
	st_case_331:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr494
		case 13:
			goto tr335
		case 32:
			goto tr493
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st332
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr493
		}
		goto st32
	st332:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof332
		}
	st_case_332:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr494
		case 13:
			goto tr335
		case 32:
			goto tr493
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st333
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr493
		}
		goto st32
	st333:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof333
		}
	st_case_333:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr494
		case 13:
			goto tr335
		case 32:
			goto tr493
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st334
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr493
		}
		goto st32
	st334:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof334
		}
	st_case_334:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr494
		case 13:
			goto tr335
		case 32:
			goto tr493
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st335
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr493
		}
		goto st32
	st335:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof335
		}
	st_case_335:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr494
		case 13:
			goto tr335
		case 32:
			goto tr493
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st336
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr493
		}
		goto st32
	st336:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof336
		}
	st_case_336:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr494
		case 13:
			goto tr335
		case 32:
			goto tr493
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st337
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr493
		}
		goto st32
	st337:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof337
		}
	st_case_337:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr494
		case 13:
			goto tr335
		case 32:
			goto tr493
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st338
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr493
		}
		goto st32
	st338:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof338
		}
	st_case_338:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr494
		case 13:
			goto tr335
		case 32:
			goto tr493
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st339
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr493
		}
		goto st32
	st339:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof339
		}
	st_case_339:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr494
		case 13:
			goto tr335
		case 32:
			goto tr493
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st340
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr493
		}
		goto st32
	st340:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof340
		}
	st_case_340:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr494
		case 13:
			goto tr335
		case 32:
			goto tr493
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st341
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr493
		}
		goto st32
	st341:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof341
		}
	st_case_341:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr494
		case 13:
			goto tr335
		case 32:
			goto tr493
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st342
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr493
		}
		goto st32
	st342:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof342
		}
	st_case_342:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr494
		case 13:
			goto tr335
		case 32:
			goto tr493
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st343
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr493
		}
		goto st32
	st343:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof343
		}
	st_case_343:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr494
		case 13:
			goto tr335
		case 32:
			goto tr493
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st344
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr493
		}
		goto st32
	st344:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof344
		}
	st_case_344:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr494
		case 13:
			goto tr335
		case 32:
			goto tr493
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st345
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr493
		}
		goto st32
	st345:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof345
		}
	st_case_345:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr494
		case 13:
			goto tr335
		case 32:
			goto tr493
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st346
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr493
		}
		goto st32
	st346:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof346
		}
	st_case_346:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr494
		case 13:
			goto tr335
		case 32:
			goto tr493
		case 44:
			goto tr61
		case 61:
			goto tr14
		case 92:
			goto st34
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr493
		}
		goto st32
tr184:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st96
tr178:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st96
	st96:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof96
		}
	st_case_96:
//line plugins/parsers/influx/machine.go:10398
		switch ( m.data)[( m.p)] {
		case 9:
			goto st7
		case 10:
			goto tr59
		case 32:
			goto st7
		case 34:
			goto tr204
		case 44:
			goto st7
		case 61:
			goto st7
		case 92:
			goto tr205
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto tr203
tr203:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st97
	st97:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof97
		}
	st_case_97:
//line plugins/parsers/influx/machine.go:10430
		switch ( m.data)[( m.p)] {
		case 9:
			goto st7
		case 10:
			goto tr59
		case 32:
			goto st7
		case 34:
			goto tr207
		case 44:
			goto st7
		case 61:
			goto tr208
		case 92:
			goto st101
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st97
tr204:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st347
tr207:
//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st347
	st347:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof347
		}
	st_case_347:
//line plugins/parsers/influx/machine.go:10472
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto st348
		case 13:
			goto tr330
		case 32:
			goto st192
		case 44:
			goto st9
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st192
		}
		goto st28
	st348:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof348
		}
	st_case_348:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto st348
		case 13:
			goto tr330
		case 32:
			goto st192
		case 44:
			goto tr210
		case 45:
			goto tr516
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr517
			}
		case ( m.data)[( m.p)] >= 9:
			goto st192
		}
		goto st28
tr516:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st98
	st98:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof98
		}
	st_case_98:
//line plugins/parsers/influx/machine.go:10536
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr210
		case 44:
			goto tr210
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] < 12:
			if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 10 {
				goto tr210
			}
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st349
			}
		default:
			goto tr210
		}
		goto st28
tr517:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st349
	st349:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof349
		}
	st_case_349:
//line plugins/parsers/influx/machine.go:10571
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr518
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr210
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st351
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st28
tr518:
//line plugins/parsers/influx/machine.go.rl:108

	m.handler.SetTimestamp(m.text())

	goto st350
	st350:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof350
		}
	st_case_350:
//line plugins/parsers/influx/machine.go:10608
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto st350
		case 13:
			goto tr330
		case 32:
			goto st195
		case 44:
			goto tr50
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st195
		}
		goto st28
	st351:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof351
		}
	st_case_351:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr518
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr210
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st352
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st28
	st352:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof352
		}
	st_case_352:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr518
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr210
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st353
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st28
	st353:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof353
		}
	st_case_353:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr518
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr210
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st354
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st28
	st354:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof354
		}
	st_case_354:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr518
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr210
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st355
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st28
	st355:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof355
		}
	st_case_355:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr518
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr210
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st356
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st28
	st356:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof356
		}
	st_case_356:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr518
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr210
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st357
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st28
	st357:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof357
		}
	st_case_357:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr518
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr210
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st358
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st28
	st358:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof358
		}
	st_case_358:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr518
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr210
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st359
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st28
	st359:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof359
		}
	st_case_359:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr518
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr210
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st360
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st28
	st360:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof360
		}
	st_case_360:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr518
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr210
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st361
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st28
	st361:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof361
		}
	st_case_361:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr518
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr210
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st362
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st28
	st362:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof362
		}
	st_case_362:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr518
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr210
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st363
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st28
	st363:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof363
		}
	st_case_363:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr518
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr210
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st364
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st28
	st364:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof364
		}
	st_case_364:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr518
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr210
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st365
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st28
	st365:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof365
		}
	st_case_365:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr518
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr210
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st366
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st28
	st366:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof366
		}
	st_case_366:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr518
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr210
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st367
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st28
	st367:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof367
		}
	st_case_367:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr518
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr210
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st368
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st28
	st368:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof368
		}
	st_case_368:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr518
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr210
		case 61:
			goto tr53
		case 92:
			goto st36
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr334
		}
		goto st28
tr208:
//line plugins/parsers/influx/machine.go.rl:76

	key = m.text()

	goto st99
	st99:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof99
		}
	st_case_99:
//line plugins/parsers/influx/machine.go:11175
		switch ( m.data)[( m.p)] {
		case 9:
			goto st7
		case 10:
			goto tr59
		case 32:
			goto st7
		case 34:
			goto tr177
		case 44:
			goto st7
		case 61:
			goto st7
		case 92:
			goto tr179
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto tr174
tr177:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st369
tr183:
//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st369
	st369:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof369
		}
	st_case_369:
//line plugins/parsers/influx/machine.go:11217
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto tr538
		case 13:
			goto tr330
		case 32:
			goto tr487
		case 44:
			goto tr489
		case 61:
			goto tr201
		case 92:
			goto st35
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr487
		}
		goto st30
tr538:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st370
tr540:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

	goto st370
tr546:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

	goto st370
tr550:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

	goto st370
tr554:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

	goto st370
	st370:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof370
		}
	st_case_370:
//line plugins/parsers/influx/machine.go:11289
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto tr490
		case 13:
			goto tr330
		case 32:
			goto tr487
		case 44:
			goto tr61
		case 45:
			goto tr491
		case 61:
			goto tr201
		case 92:
			goto tr65
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr492
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr487
		}
		goto tr63
tr179:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st100
	st100:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof100
		}
	st_case_100:
//line plugins/parsers/influx/machine.go:11328
		switch ( m.data)[( m.p)] {
		case 34:
			goto st86
		case 92:
			goto st86
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr59
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr59
		}
		goto st30
tr205:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st101
	st101:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof101
		}
	st_case_101:
//line plugins/parsers/influx/machine.go:11355
		switch ( m.data)[( m.p)] {
		case 34:
			goto st97
		case 92:
			goto st97
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr59
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr59
		}
		goto st28
tr196:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st102
	st102:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof102
		}
	st_case_102:
//line plugins/parsers/influx/machine.go:11382
		switch ( m.data)[( m.p)] {
		case 34:
			goto st93
		case 92:
			goto st93
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr59
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr59
		}
		goto st32
tr166:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st103
	st103:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof103
		}
	st_case_103:
//line plugins/parsers/influx/machine.go:11409
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr59
		case 11:
			goto tr60
		case 13:
			goto tr59
		case 32:
			goto tr58
		case 44:
			goto tr61
		case 48:
			goto st371
		case 61:
			goto tr59
		case 92:
			goto st35
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st375
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr58
		}
		goto st30
	st371:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof371
		}
	st_case_371:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 11:
			goto tr540
		case 13:
			goto tr356
		case 32:
			goto tr539
		case 44:
			goto tr541
		case 46:
			goto st372
		case 61:
			goto tr201
		case 69:
			goto st104
		case 92:
			goto st35
		case 101:
			goto st104
		case 105:
			goto st374
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr539
		}
		goto st30
tr167:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st372
	st372:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof372
		}
	st_case_372:
//line plugins/parsers/influx/machine.go:11481
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 11:
			goto tr540
		case 13:
			goto tr356
		case 32:
			goto tr539
		case 44:
			goto tr541
		case 61:
			goto tr201
		case 69:
			goto st104
		case 92:
			goto st35
		case 101:
			goto st104
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st372
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr539
		}
		goto st30
	st104:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof104
		}
	st_case_104:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr59
		case 11:
			goto tr60
		case 13:
			goto tr59
		case 32:
			goto tr58
		case 34:
			goto st105
		case 44:
			goto tr61
		case 61:
			goto tr59
		case 92:
			goto st35
		}
		switch {
		case ( m.data)[( m.p)] < 43:
			if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
				goto tr58
			}
		case ( m.data)[( m.p)] > 45:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st373
			}
		default:
			goto st105
		}
		goto st30
	st105:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof105
		}
	st_case_105:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr59
		case 11:
			goto tr60
		case 13:
			goto tr59
		case 32:
			goto tr58
		case 44:
			goto tr61
		case 61:
			goto tr59
		case 92:
			goto st35
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st373
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr58
		}
		goto st30
	st373:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof373
		}
	st_case_373:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 11:
			goto tr540
		case 13:
			goto tr356
		case 32:
			goto tr539
		case 44:
			goto tr541
		case 61:
			goto tr201
		case 92:
			goto st35
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st373
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr539
		}
		goto st30
	st374:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof374
		}
	st_case_374:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr362
		case 11:
			goto tr546
		case 13:
			goto tr362
		case 32:
			goto tr545
		case 44:
			goto tr547
		case 61:
			goto tr201
		case 92:
			goto st35
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr545
		}
		goto st30
	st375:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof375
		}
	st_case_375:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 11:
			goto tr540
		case 13:
			goto tr356
		case 32:
			goto tr539
		case 44:
			goto tr541
		case 46:
			goto st372
		case 61:
			goto tr201
		case 69:
			goto st104
		case 92:
			goto st35
		case 101:
			goto st104
		case 105:
			goto st374
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st375
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr539
		}
		goto st30
tr168:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st376
	st376:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof376
		}
	st_case_376:
//line plugins/parsers/influx/machine.go:11681
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 11:
			goto tr540
		case 13:
			goto tr356
		case 32:
			goto tr539
		case 44:
			goto tr541
		case 46:
			goto st372
		case 61:
			goto tr201
		case 69:
			goto st104
		case 92:
			goto st35
		case 101:
			goto st104
		case 105:
			goto st374
		case 117:
			goto st377
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr539
		}
		goto st30
	st377:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof377
		}
	st_case_377:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr366
		case 11:
			goto tr550
		case 13:
			goto tr366
		case 32:
			goto tr549
		case 44:
			goto tr551
		case 61:
			goto tr201
		case 92:
			goto st35
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr549
		}
		goto st30
tr169:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st378
	st378:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof378
		}
	st_case_378:
//line plugins/parsers/influx/machine.go:11748
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 11:
			goto tr540
		case 13:
			goto tr356
		case 32:
			goto tr539
		case 44:
			goto tr541
		case 46:
			goto st372
		case 61:
			goto tr201
		case 69:
			goto st104
		case 92:
			goto st35
		case 101:
			goto st104
		case 105:
			goto st374
		case 117:
			goto st377
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st378
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr539
		}
		goto st30
tr170:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st379
	st379:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof379
		}
	st_case_379:
//line plugins/parsers/influx/machine.go:11795
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 11:
			goto tr554
		case 13:
			goto tr370
		case 32:
			goto tr553
		case 44:
			goto tr555
		case 61:
			goto tr201
		case 65:
			goto st106
		case 92:
			goto st35
		case 97:
			goto st109
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr553
		}
		goto st30
	st106:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof106
		}
	st_case_106:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr59
		case 11:
			goto tr60
		case 13:
			goto tr59
		case 32:
			goto tr58
		case 44:
			goto tr61
		case 61:
			goto tr59
		case 76:
			goto st107
		case 92:
			goto st35
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr58
		}
		goto st30
	st107:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof107
		}
	st_case_107:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr59
		case 11:
			goto tr60
		case 13:
			goto tr59
		case 32:
			goto tr58
		case 44:
			goto tr61
		case 61:
			goto tr59
		case 83:
			goto st108
		case 92:
			goto st35
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr58
		}
		goto st30
	st108:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof108
		}
	st_case_108:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr59
		case 11:
			goto tr60
		case 13:
			goto tr59
		case 32:
			goto tr58
		case 44:
			goto tr61
		case 61:
			goto tr59
		case 69:
			goto st380
		case 92:
			goto st35
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr58
		}
		goto st30
	st380:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof380
		}
	st_case_380:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 11:
			goto tr554
		case 13:
			goto tr370
		case 32:
			goto tr553
		case 44:
			goto tr555
		case 61:
			goto tr201
		case 92:
			goto st35
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr553
		}
		goto st30
	st109:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof109
		}
	st_case_109:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr59
		case 11:
			goto tr60
		case 13:
			goto tr59
		case 32:
			goto tr58
		case 44:
			goto tr61
		case 61:
			goto tr59
		case 92:
			goto st35
		case 108:
			goto st110
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr58
		}
		goto st30
	st110:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof110
		}
	st_case_110:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr59
		case 11:
			goto tr60
		case 13:
			goto tr59
		case 32:
			goto tr58
		case 44:
			goto tr61
		case 61:
			goto tr59
		case 92:
			goto st35
		case 115:
			goto st111
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr58
		}
		goto st30
	st111:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof111
		}
	st_case_111:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr59
		case 11:
			goto tr60
		case 13:
			goto tr59
		case 32:
			goto tr58
		case 44:
			goto tr61
		case 61:
			goto tr59
		case 92:
			goto st35
		case 101:
			goto st380
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr58
		}
		goto st30
tr171:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st381
	st381:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof381
		}
	st_case_381:
//line plugins/parsers/influx/machine.go:12018
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 11:
			goto tr554
		case 13:
			goto tr370
		case 32:
			goto tr553
		case 44:
			goto tr555
		case 61:
			goto tr201
		case 82:
			goto st112
		case 92:
			goto st35
		case 114:
			goto st113
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr553
		}
		goto st30
	st112:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof112
		}
	st_case_112:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr59
		case 11:
			goto tr60
		case 13:
			goto tr59
		case 32:
			goto tr58
		case 44:
			goto tr61
		case 61:
			goto tr59
		case 85:
			goto st108
		case 92:
			goto st35
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr58
		}
		goto st30
	st113:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof113
		}
	st_case_113:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr59
		case 11:
			goto tr60
		case 13:
			goto tr59
		case 32:
			goto tr58
		case 44:
			goto tr61
		case 61:
			goto tr59
		case 92:
			goto st35
		case 117:
			goto st111
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr58
		}
		goto st30
tr172:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st382
	st382:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof382
		}
	st_case_382:
//line plugins/parsers/influx/machine.go:12108
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 11:
			goto tr554
		case 13:
			goto tr370
		case 32:
			goto tr553
		case 44:
			goto tr555
		case 61:
			goto tr201
		case 92:
			goto st35
		case 97:
			goto st109
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr553
		}
		goto st30
tr173:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st383
	st383:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof383
		}
	st_case_383:
//line plugins/parsers/influx/machine.go:12142
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 11:
			goto tr554
		case 13:
			goto tr370
		case 32:
			goto tr553
		case 44:
			goto tr555
		case 61:
			goto tr201
		case 92:
			goto st35
		case 114:
			goto st113
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr553
		}
		goto st30
tr161:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st114
	st114:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof114
		}
	st_case_114:
//line plugins/parsers/influx/machine.go:12176
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr59
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr59
		}
		goto st83
tr88:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

	goto st115
tr82:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st115
tr231:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st115
	st115:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof115
		}
	st_case_115:
//line plugins/parsers/influx/machine.go:12213
		switch ( m.data)[( m.p)] {
		case 9:
			goto st7
		case 10:
			goto tr59
		case 32:
			goto st7
		case 34:
			goto tr204
		case 44:
			goto st7
		case 61:
			goto st7
		case 92:
			goto tr222
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto tr221
tr221:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st116
	st116:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof116
		}
	st_case_116:
//line plugins/parsers/influx/machine.go:12245
		switch ( m.data)[( m.p)] {
		case 9:
			goto st7
		case 10:
			goto tr59
		case 32:
			goto st7
		case 34:
			goto tr207
		case 44:
			goto st7
		case 61:
			goto tr224
		case 92:
			goto st124
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st116
tr224:
//line plugins/parsers/influx/machine.go.rl:76

	key = m.text()

	goto st117
	st117:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof117
		}
	st_case_117:
//line plugins/parsers/influx/machine.go:12277
		switch ( m.data)[( m.p)] {
		case 9:
			goto st7
		case 10:
			goto tr59
		case 32:
			goto st7
		case 34:
			goto tr177
		case 44:
			goto st7
		case 61:
			goto st7
		case 92:
			goto tr227
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto tr226
tr226:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st118
	st118:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof118
		}
	st_case_118:
//line plugins/parsers/influx/machine.go:12309
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr229
		case 11:
			goto tr230
		case 12:
			goto tr58
		case 32:
			goto tr229
		case 34:
			goto tr183
		case 44:
			goto tr231
		case 61:
			goto st7
		case 92:
			goto st123
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st118
tr230:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st119
	st119:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof119
		}
	st_case_119:
//line plugins/parsers/influx/machine.go:12343
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr229
		case 11:
			goto tr234
		case 12:
			goto tr58
		case 32:
			goto tr229
		case 34:
			goto tr195
		case 44:
			goto tr231
		case 61:
			goto st7
		case 92:
			goto tr235
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto tr233
tr233:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st120
	st120:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof120
		}
	st_case_120:
//line plugins/parsers/influx/machine.go:12377
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr229
		case 11:
			goto tr237
		case 12:
			goto tr58
		case 32:
			goto tr229
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st120
tr237:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st121
tr234:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st121
	st121:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof121
		}
	st_case_121:
//line plugins/parsers/influx/machine.go:12421
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr229
		case 11:
			goto tr234
		case 12:
			goto tr58
		case 32:
			goto tr229
		case 34:
			goto tr195
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto tr235
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto tr233
tr235:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st122
	st122:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof122
		}
	st_case_122:
//line plugins/parsers/influx/machine.go:12455
		switch ( m.data)[( m.p)] {
		case 34:
			goto st120
		case 92:
			goto st120
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr59
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr59
		}
		goto st32
tr227:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st123
	st123:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof123
		}
	st_case_123:
//line plugins/parsers/influx/machine.go:12482
		switch ( m.data)[( m.p)] {
		case 34:
			goto st118
		case 92:
			goto st118
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr59
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr59
		}
		goto st30
tr222:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st124
	st124:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof124
		}
	st_case_124:
//line plugins/parsers/influx/machine.go:12509
		switch ( m.data)[( m.p)] {
		case 34:
			goto st116
		case 92:
			goto st116
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr59
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr59
		}
		goto st28
tr157:
//line plugins/parsers/influx/machine.go.rl:84

	key = m.text()

	goto st125
	st125:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof125
		}
	st_case_125:
//line plugins/parsers/influx/machine.go:12536
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr85
		case 11:
			goto tr86
		case 12:
			goto tr4
		case 32:
			goto tr85
		case 34:
			goto tr239
		case 44:
			goto tr88
		case 45:
			goto tr240
		case 46:
			goto tr241
		case 48:
			goto tr242
		case 70:
			goto tr244
		case 84:
			goto tr245
		case 92:
			goto st164
		case 102:
			goto tr246
		case 116:
			goto tr247
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr243
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr5
		}
		goto st39
tr239:
//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st384
	st384:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof384
		}
	st_case_384:
//line plugins/parsers/influx/machine.go:12587
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr560
		case 11:
			goto tr561
		case 12:
			goto tr455
		case 32:
			goto tr560
		case 34:
			goto tr81
		case 44:
			goto tr562
		case 92:
			goto tr83
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr330
		}
		goto tr78
tr587:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

	goto st385
tr560:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st385
tr721:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st385
tr715:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

	goto st385
tr747:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

	goto st385
tr751:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

	goto st385
tr755:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

	goto st385
tr762:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

	goto st385
tr771:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

	goto st385
tr776:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

	goto st385
tr781:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

	goto st385
	st385:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof385
		}
	st_case_385:
//line plugins/parsers/influx/machine.go:12715
		switch ( m.data)[( m.p)] {
		case 9:
			goto st385
		case 11:
			goto tr564
		case 12:
			goto st300
		case 32:
			goto st385
		case 34:
			goto tr93
		case 44:
			goto st7
		case 45:
			goto tr565
		case 61:
			goto st7
		case 92:
			goto tr94
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr566
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr330
		}
		goto tr90
tr564:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st386
	st386:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof386
		}
	st_case_386:
//line plugins/parsers/influx/machine.go:12756
		switch ( m.data)[( m.p)] {
		case 9:
			goto st385
		case 11:
			goto tr564
		case 12:
			goto st300
		case 32:
			goto st385
		case 34:
			goto tr93
		case 44:
			goto st7
		case 45:
			goto tr565
		case 61:
			goto tr97
		case 92:
			goto tr94
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr566
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr330
		}
		goto tr90
tr565:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st126
	st126:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof126
		}
	st_case_126:
//line plugins/parsers/influx/machine.go:12797
		switch ( m.data)[( m.p)] {
		case 9:
			goto st7
		case 10:
			goto tr99
		case 32:
			goto st7
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st387
			}
		case ( m.data)[( m.p)] >= 12:
			goto tr99
		}
		goto st41
tr566:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st387
	st387:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof387
		}
	st_case_387:
//line plugins/parsers/influx/machine.go:12834
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr567
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st389
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st41
tr567:
//line plugins/parsers/influx/machine.go.rl:108

	m.handler.SetTimestamp(m.text())

	goto st388
	st388:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof388
		}
	st_case_388:
//line plugins/parsers/influx/machine.go:12873
		switch ( m.data)[( m.p)] {
		case 9:
			goto st252
		case 11:
			goto st388
		case 12:
			goto st195
		case 32:
			goto st252
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr330
		}
		goto st41
	st389:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof389
		}
	st_case_389:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr567
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st390
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st41
	st390:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof390
		}
	st_case_390:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr567
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st391
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st41
	st391:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof391
		}
	st_case_391:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr567
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st392
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st41
	st392:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof392
		}
	st_case_392:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr567
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st393
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st41
	st393:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof393
		}
	st_case_393:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr567
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st394
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st41
	st394:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof394
		}
	st_case_394:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr567
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st395
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st41
	st395:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof395
		}
	st_case_395:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr567
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st396
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st41
	st396:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof396
		}
	st_case_396:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr567
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st397
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st41
	st397:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof397
		}
	st_case_397:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr567
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st398
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st41
	st398:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof398
		}
	st_case_398:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr567
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st399
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st41
	st399:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof399
		}
	st_case_399:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr567
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st400
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st41
	st400:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof400
		}
	st_case_400:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr567
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st401
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st41
	st401:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof401
		}
	st_case_401:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr567
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st402
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st41
	st402:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof402
		}
	st_case_402:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr567
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st403
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st41
	st403:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof403
		}
	st_case_403:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr567
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st404
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st41
	st404:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof404
		}
	st_case_404:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr567
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st405
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st41
	st405:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof405
		}
	st_case_405:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr567
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st406
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st41
	st406:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof406
		}
	st_case_406:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr567
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto st75
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr335
		}
		goto st41
tr561:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st407
tr763:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

	goto st407
tr772:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

	goto st407
tr777:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

	goto st407
tr782:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

	goto st407
	st407:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof407
		}
	st_case_407:
//line plugins/parsers/influx/machine.go:13522
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr587
		case 11:
			goto tr588
		case 12:
			goto tr455
		case 32:
			goto tr587
		case 34:
			goto tr152
		case 44:
			goto tr88
		case 45:
			goto tr589
		case 61:
			goto st39
		case 92:
			goto tr153
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr590
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr330
		}
		goto tr150
tr588:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st408
	st408:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof408
		}
	st_case_408:
//line plugins/parsers/influx/machine.go:13567
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr587
		case 11:
			goto tr588
		case 12:
			goto tr455
		case 32:
			goto tr587
		case 34:
			goto tr152
		case 44:
			goto tr88
		case 45:
			goto tr589
		case 61:
			goto tr157
		case 92:
			goto tr153
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr590
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr330
		}
		goto tr150
tr589:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st127
	st127:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof127
		}
	st_case_127:
//line plugins/parsers/influx/machine.go:13608
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr85
		case 11:
			goto tr155
		case 12:
			goto tr4
		case 32:
			goto tr85
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st409
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr99
		}
		goto st78
tr590:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st409
	st409:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof409
		}
	st_case_409:
//line plugins/parsers/influx/machine.go:13647
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr591
		case 11:
			goto tr592
		case 12:
			goto tr463
		case 32:
			goto tr591
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st413
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st78
tr596:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

	goto st410
tr728:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st410
tr591:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:108

	m.handler.SetTimestamp(m.text())

	goto st410
tr725:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:108

	m.handler.SetTimestamp(m.text())

	goto st410
	st410:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof410
		}
	st_case_410:
//line plugins/parsers/influx/machine.go:13712
		switch ( m.data)[( m.p)] {
		case 9:
			goto st410
		case 11:
			goto tr595
		case 12:
			goto st304
		case 32:
			goto st410
		case 34:
			goto tr93
		case 44:
			goto st7
		case 61:
			goto st7
		case 92:
			goto tr94
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr330
		}
		goto tr90
tr595:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st411
	st411:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof411
		}
	st_case_411:
//line plugins/parsers/influx/machine.go:13746
		switch ( m.data)[( m.p)] {
		case 9:
			goto st410
		case 11:
			goto tr595
		case 12:
			goto st304
		case 32:
			goto st410
		case 34:
			goto tr93
		case 44:
			goto st7
		case 61:
			goto tr97
		case 92:
			goto tr94
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr330
		}
		goto tr90
tr597:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st412
tr592:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:108

	m.handler.SetTimestamp(m.text())

	goto st412
	st412:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof412
		}
	st_case_412:
//line plugins/parsers/influx/machine.go:13794
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr596
		case 11:
			goto tr597
		case 12:
			goto tr468
		case 32:
			goto tr596
		case 34:
			goto tr152
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto tr153
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr330
		}
		goto tr150
tr153:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st128
	st128:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof128
		}
	st_case_128:
//line plugins/parsers/influx/machine.go:13828
		switch ( m.data)[( m.p)] {
		case 34:
			goto st78
		case 92:
			goto st78
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr5
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr5
		}
		goto st25
	st413:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof413
		}
	st_case_413:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr591
		case 11:
			goto tr592
		case 12:
			goto tr463
		case 32:
			goto tr591
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st414
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st78
	st414:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof414
		}
	st_case_414:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr591
		case 11:
			goto tr592
		case 12:
			goto tr463
		case 32:
			goto tr591
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st415
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st78
	st415:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof415
		}
	st_case_415:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr591
		case 11:
			goto tr592
		case 12:
			goto tr463
		case 32:
			goto tr591
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st416
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st78
	st416:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof416
		}
	st_case_416:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr591
		case 11:
			goto tr592
		case 12:
			goto tr463
		case 32:
			goto tr591
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st417
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st78
	st417:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof417
		}
	st_case_417:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr591
		case 11:
			goto tr592
		case 12:
			goto tr463
		case 32:
			goto tr591
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st418
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st78
	st418:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof418
		}
	st_case_418:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr591
		case 11:
			goto tr592
		case 12:
			goto tr463
		case 32:
			goto tr591
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st419
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st78
	st419:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof419
		}
	st_case_419:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr591
		case 11:
			goto tr592
		case 12:
			goto tr463
		case 32:
			goto tr591
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st420
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st78
	st420:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof420
		}
	st_case_420:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr591
		case 11:
			goto tr592
		case 12:
			goto tr463
		case 32:
			goto tr591
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st421
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st78
	st421:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof421
		}
	st_case_421:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr591
		case 11:
			goto tr592
		case 12:
			goto tr463
		case 32:
			goto tr591
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st422
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st78
	st422:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof422
		}
	st_case_422:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr591
		case 11:
			goto tr592
		case 12:
			goto tr463
		case 32:
			goto tr591
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st423
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st78
	st423:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof423
		}
	st_case_423:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr591
		case 11:
			goto tr592
		case 12:
			goto tr463
		case 32:
			goto tr591
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st424
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st78
	st424:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof424
		}
	st_case_424:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr591
		case 11:
			goto tr592
		case 12:
			goto tr463
		case 32:
			goto tr591
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st425
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st78
	st425:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof425
		}
	st_case_425:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr591
		case 11:
			goto tr592
		case 12:
			goto tr463
		case 32:
			goto tr591
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st426
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st78
	st426:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof426
		}
	st_case_426:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr591
		case 11:
			goto tr592
		case 12:
			goto tr463
		case 32:
			goto tr591
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st427
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st78
	st427:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof427
		}
	st_case_427:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr591
		case 11:
			goto tr592
		case 12:
			goto tr463
		case 32:
			goto tr591
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st428
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st78
	st428:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof428
		}
	st_case_428:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr591
		case 11:
			goto tr592
		case 12:
			goto tr463
		case 32:
			goto tr591
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st429
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st78
	st429:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof429
		}
	st_case_429:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr591
		case 11:
			goto tr592
		case 12:
			goto tr463
		case 32:
			goto tr591
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st430
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st78
	st430:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof430
		}
	st_case_430:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr591
		case 11:
			goto tr592
		case 12:
			goto tr463
		case 32:
			goto tr591
		case 34:
			goto tr156
		case 44:
			goto tr88
		case 61:
			goto tr157
		case 92:
			goto st128
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr335
		}
		goto st78
tr81:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st431
tr87:
//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st431
	st431:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof431
		}
	st_case_431:
//line plugins/parsers/influx/machine.go:14436
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto tr615
		case 13:
			goto tr330
		case 32:
			goto tr455
		case 44:
			goto tr457
		case 92:
			goto st129
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr455
		}
		goto st2
tr615:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

	goto st432
tr769:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

	goto st432
tr792:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

	goto st432
tr795:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

	goto st432
tr798:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

	goto st432
	st432:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof432
		}
	st_case_432:
//line plugins/parsers/influx/machine.go:14506
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto tr460
		case 13:
			goto tr330
		case 32:
			goto tr455
		case 44:
			goto tr7
		case 45:
			goto tr461
		case 61:
			goto st2
		case 92:
			goto tr44
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr462
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr455
		}
		goto tr42
tr2:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st129
	st129:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof129
		}
	st_case_129:
//line plugins/parsers/influx/machine.go:14545
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr1
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr1
		}
		goto st2
tr562:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st130
tr717:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

	goto st130
tr749:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

	goto st130
tr753:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

	goto st130
tr757:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

	goto st130
tr765:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

	goto st130
tr774:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

	goto st130
tr779:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

	goto st130
tr784:
//line plugins/parsers/influx/machine.go.rl:72

	m.handler.SetMeasurement(m.text())

//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

	goto st130
	st130:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof130
		}
	st_case_130:
//line plugins/parsers/influx/machine.go:14650
		switch ( m.data)[( m.p)] {
		case 9:
			goto st7
		case 10:
			goto tr59
		case 32:
			goto st7
		case 34:
			goto tr251
		case 44:
			goto st7
		case 61:
			goto st7
		case 92:
			goto tr252
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto tr250
tr250:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st131
	st131:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof131
		}
	st_case_131:
//line plugins/parsers/influx/machine.go:14682
		switch ( m.data)[( m.p)] {
		case 9:
			goto st7
		case 10:
			goto tr59
		case 32:
			goto st7
		case 34:
			goto tr254
		case 44:
			goto st7
		case 61:
			goto tr255
		case 92:
			goto st163
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st131
tr251:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st433
tr254:
//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st433
	st433:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof433
		}
	st_case_433:
//line plugins/parsers/influx/machine.go:14724
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto st434
		case 13:
			goto tr330
		case 32:
			goto st192
		case 44:
			goto st9
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st192
		}
		goto st83
	st434:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof434
		}
	st_case_434:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto st434
		case 13:
			goto tr330
		case 32:
			goto st192
		case 44:
			goto tr201
		case 45:
			goto tr617
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr618
			}
		case ( m.data)[( m.p)] >= 9:
			goto st192
		}
		goto st83
tr617:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st132
	st132:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof132
		}
	st_case_132:
//line plugins/parsers/influx/machine.go:14788
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr201
		case 44:
			goto tr201
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] < 12:
			if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 10 {
				goto tr201
			}
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st435
			}
		default:
			goto tr201
		}
		goto st83
tr618:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st435
	st435:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof435
		}
	st_case_435:
//line plugins/parsers/influx/machine.go:14823
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr619
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr201
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st437
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st83
tr619:
//line plugins/parsers/influx/machine.go.rl:108

	m.handler.SetTimestamp(m.text())

	goto st436
	st436:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof436
		}
	st_case_436:
//line plugins/parsers/influx/machine.go:14860
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto st436
		case 13:
			goto tr330
		case 32:
			goto st195
		case 44:
			goto tr59
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st195
		}
		goto st83
	st437:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof437
		}
	st_case_437:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr619
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr201
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st438
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st83
	st438:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof438
		}
	st_case_438:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr619
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr201
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st439
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st83
	st439:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof439
		}
	st_case_439:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr619
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr201
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st440
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st83
	st440:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof440
		}
	st_case_440:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr619
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr201
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st441
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st83
	st441:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof441
		}
	st_case_441:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr619
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr201
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st442
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st83
	st442:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof442
		}
	st_case_442:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr619
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr201
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st443
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st83
	st443:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof443
		}
	st_case_443:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr619
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr201
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st444
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st83
	st444:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof444
		}
	st_case_444:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr619
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr201
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st445
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st83
	st445:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof445
		}
	st_case_445:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr619
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr201
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st446
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st83
	st446:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof446
		}
	st_case_446:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr619
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr201
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st447
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st83
	st447:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof447
		}
	st_case_447:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr619
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr201
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st448
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st83
	st448:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof448
		}
	st_case_448:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr619
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr201
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st449
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st83
	st449:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof449
		}
	st_case_449:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr619
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr201
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st450
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st83
	st450:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof450
		}
	st_case_450:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr619
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr201
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st451
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st83
	st451:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof451
		}
	st_case_451:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr619
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr201
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st452
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st83
	st452:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof452
		}
	st_case_452:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr619
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr201
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st453
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st83
	st453:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof453
		}
	st_case_453:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr619
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr201
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st454
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr334
		}
		goto st83
	st454:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof454
		}
	st_case_454:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr335
		case 11:
			goto tr619
		case 13:
			goto tr335
		case 32:
			goto tr334
		case 44:
			goto tr201
		case 61:
			goto tr163
		case 92:
			goto st114
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr334
		}
		goto st83
tr255:
//line plugins/parsers/influx/machine.go.rl:76

	key = m.text()

//line plugins/parsers/influx/machine.go.rl:84

	key = m.text()

	goto st133
	st133:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof133
		}
	st_case_133:
//line plugins/parsers/influx/machine.go:15431
		switch ( m.data)[( m.p)] {
		case 9:
			goto st7
		case 10:
			goto tr59
		case 32:
			goto st7
		case 34:
			goto tr258
		case 44:
			goto st7
		case 45:
			goto tr259
		case 46:
			goto tr260
		case 48:
			goto tr261
		case 61:
			goto st7
		case 70:
			goto tr263
		case 84:
			goto tr264
		case 92:
			goto tr227
		case 102:
			goto tr265
		case 116:
			goto tr266
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr262
			}
		case ( m.data)[( m.p)] >= 12:
			goto tr59
		}
		goto tr226
tr258:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st455
	st455:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof455
		}
	st_case_455:
//line plugins/parsers/influx/machine.go:15486
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr639
		case 11:
			goto tr640
		case 12:
			goto tr487
		case 32:
			goto tr639
		case 34:
			goto tr177
		case 44:
			goto tr641
		case 61:
			goto tr25
		case 92:
			goto tr179
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr330
		}
		goto tr174
tr666:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st456
tr639:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st456
tr694:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

	goto st456
tr700:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

	goto st456
tr704:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

	goto st456
tr708:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

	goto st456
	st456:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof456
		}
	st_case_456:
//line plugins/parsers/influx/machine.go:15570
		switch ( m.data)[( m.p)] {
		case 9:
			goto st456
		case 11:
			goto tr643
		case 12:
			goto st300
		case 32:
			goto st456
		case 34:
			goto tr93
		case 44:
			goto st7
		case 45:
			goto tr644
		case 61:
			goto st7
		case 92:
			goto tr189
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr645
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr330
		}
		goto tr186
tr643:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st457
	st457:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof457
		}
	st_case_457:
//line plugins/parsers/influx/machine.go:15611
		switch ( m.data)[( m.p)] {
		case 9:
			goto st456
		case 11:
			goto tr643
		case 12:
			goto st300
		case 32:
			goto st456
		case 34:
			goto tr93
		case 44:
			goto st7
		case 45:
			goto tr644
		case 61:
			goto tr191
		case 92:
			goto tr189
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr645
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr330
		}
		goto tr186
tr644:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st134
	st134:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof134
		}
	st_case_134:
//line plugins/parsers/influx/machine.go:15652
		switch ( m.data)[( m.p)] {
		case 9:
			goto st7
		case 10:
			goto tr99
		case 32:
			goto st7
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st458
			}
		case ( m.data)[( m.p)] >= 12:
			goto tr99
		}
		goto st88
tr645:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st458
	st458:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof458
		}
	st_case_458:
//line plugins/parsers/influx/machine.go:15689
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr646
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st460
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st88
tr646:
//line plugins/parsers/influx/machine.go.rl:108

	m.handler.SetTimestamp(m.text())

	goto st459
	st459:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof459
		}
	st_case_459:
//line plugins/parsers/influx/machine.go:15728
		switch ( m.data)[( m.p)] {
		case 9:
			goto st252
		case 11:
			goto st459
		case 12:
			goto st195
		case 32:
			goto st252
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr330
		}
		goto st88
	st460:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof460
		}
	st_case_460:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr646
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st461
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st88
	st461:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof461
		}
	st_case_461:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr646
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st462
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st88
	st462:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof462
		}
	st_case_462:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr646
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st463
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st88
	st463:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof463
		}
	st_case_463:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr646
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st464
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st88
	st464:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof464
		}
	st_case_464:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr646
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st465
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st88
	st465:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof465
		}
	st_case_465:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr646
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st466
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st88
	st466:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof466
		}
	st_case_466:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr646
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st467
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st88
	st467:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof467
		}
	st_case_467:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr646
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st468
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st88
	st468:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof468
		}
	st_case_468:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr646
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st469
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st88
	st469:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof469
		}
	st_case_469:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr646
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st470
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st88
	st470:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof470
		}
	st_case_470:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr646
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st471
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st88
	st471:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof471
		}
	st_case_471:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr646
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st472
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st88
	st472:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof472
		}
	st_case_472:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr646
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st473
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st88
	st473:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof473
		}
	st_case_473:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr646
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st474
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st88
	st474:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof474
		}
	st_case_474:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr646
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st475
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st88
	st475:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof475
		}
	st_case_475:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr646
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st476
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st88
	st476:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof476
		}
	st_case_476:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr646
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st477
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st88
	st477:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof477
		}
	st_case_477:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr404
		case 11:
			goto tr646
		case 12:
			goto tr334
		case 32:
			goto tr404
		case 34:
			goto tr96
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto st90
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr335
		}
		goto st88
tr640:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st478
tr695:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

	goto st478
tr701:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

	goto st478
tr705:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

	goto st478
tr709:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

	goto st478
	st478:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof478
		}
	st_case_478:
//line plugins/parsers/influx/machine.go:16377
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr666
		case 11:
			goto tr667
		case 12:
			goto tr487
		case 32:
			goto tr666
		case 34:
			goto tr195
		case 44:
			goto tr184
		case 45:
			goto tr668
		case 61:
			goto st7
		case 92:
			goto tr196
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr669
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr330
		}
		goto tr193
tr667:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st479
	st479:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof479
		}
	st_case_479:
//line plugins/parsers/influx/machine.go:16422
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr666
		case 11:
			goto tr667
		case 12:
			goto tr487
		case 32:
			goto tr666
		case 34:
			goto tr195
		case 44:
			goto tr184
		case 45:
			goto tr668
		case 61:
			goto tr191
		case 92:
			goto tr196
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr669
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr330
		}
		goto tr193
tr668:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st135
	st135:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof135
		}
	st_case_135:
//line plugins/parsers/influx/machine.go:16463
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr181
		case 11:
			goto tr198
		case 12:
			goto tr58
		case 32:
			goto tr181
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st480
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr201
		}
		goto st93
tr669:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st480
	st480:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof480
		}
	st_case_480:
//line plugins/parsers/influx/machine.go:16502
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr670
		case 11:
			goto tr671
		case 12:
			goto tr493
		case 32:
			goto tr670
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st484
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st93
tr675:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

	goto st481
tr670:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:108

	m.handler.SetTimestamp(m.text())

	goto st481
	st481:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof481
		}
	st_case_481:
//line plugins/parsers/influx/machine.go:16551
		switch ( m.data)[( m.p)] {
		case 9:
			goto st481
		case 11:
			goto tr674
		case 12:
			goto st304
		case 32:
			goto st481
		case 34:
			goto tr93
		case 44:
			goto st7
		case 61:
			goto st7
		case 92:
			goto tr189
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr330
		}
		goto tr186
tr674:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st482
	st482:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof482
		}
	st_case_482:
//line plugins/parsers/influx/machine.go:16585
		switch ( m.data)[( m.p)] {
		case 9:
			goto st481
		case 11:
			goto tr674
		case 12:
			goto st304
		case 32:
			goto st481
		case 34:
			goto tr93
		case 44:
			goto st7
		case 61:
			goto tr191
		case 92:
			goto tr189
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr330
		}
		goto tr186
tr676:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st483
tr671:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:108

	m.handler.SetTimestamp(m.text())

	goto st483
	st483:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof483
		}
	st_case_483:
//line plugins/parsers/influx/machine.go:16633
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr675
		case 11:
			goto tr676
		case 12:
			goto tr496
		case 32:
			goto tr675
		case 34:
			goto tr195
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto tr196
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr330
		}
		goto tr193
	st484:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof484
		}
	st_case_484:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr670
		case 11:
			goto tr671
		case 12:
			goto tr493
		case 32:
			goto tr670
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st485
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st93
	st485:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof485
		}
	st_case_485:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr670
		case 11:
			goto tr671
		case 12:
			goto tr493
		case 32:
			goto tr670
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st486
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st93
	st486:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof486
		}
	st_case_486:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr670
		case 11:
			goto tr671
		case 12:
			goto tr493
		case 32:
			goto tr670
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st487
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st93
	st487:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof487
		}
	st_case_487:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr670
		case 11:
			goto tr671
		case 12:
			goto tr493
		case 32:
			goto tr670
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st488
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st93
	st488:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof488
		}
	st_case_488:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr670
		case 11:
			goto tr671
		case 12:
			goto tr493
		case 32:
			goto tr670
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st489
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st93
	st489:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof489
		}
	st_case_489:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr670
		case 11:
			goto tr671
		case 12:
			goto tr493
		case 32:
			goto tr670
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st490
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st93
	st490:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof490
		}
	st_case_490:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr670
		case 11:
			goto tr671
		case 12:
			goto tr493
		case 32:
			goto tr670
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st491
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st93
	st491:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof491
		}
	st_case_491:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr670
		case 11:
			goto tr671
		case 12:
			goto tr493
		case 32:
			goto tr670
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st492
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st93
	st492:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof492
		}
	st_case_492:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr670
		case 11:
			goto tr671
		case 12:
			goto tr493
		case 32:
			goto tr670
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st493
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st93
	st493:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof493
		}
	st_case_493:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr670
		case 11:
			goto tr671
		case 12:
			goto tr493
		case 32:
			goto tr670
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st494
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st93
	st494:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof494
		}
	st_case_494:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr670
		case 11:
			goto tr671
		case 12:
			goto tr493
		case 32:
			goto tr670
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st495
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st93
	st495:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof495
		}
	st_case_495:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr670
		case 11:
			goto tr671
		case 12:
			goto tr493
		case 32:
			goto tr670
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st496
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st93
	st496:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof496
		}
	st_case_496:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr670
		case 11:
			goto tr671
		case 12:
			goto tr493
		case 32:
			goto tr670
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st497
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st93
	st497:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof497
		}
	st_case_497:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr670
		case 11:
			goto tr671
		case 12:
			goto tr493
		case 32:
			goto tr670
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st498
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st93
	st498:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof498
		}
	st_case_498:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr670
		case 11:
			goto tr671
		case 12:
			goto tr493
		case 32:
			goto tr670
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st499
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st93
	st499:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof499
		}
	st_case_499:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr670
		case 11:
			goto tr671
		case 12:
			goto tr493
		case 32:
			goto tr670
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st500
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st93
	st500:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof500
		}
	st_case_500:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr670
		case 11:
			goto tr671
		case 12:
			goto tr493
		case 32:
			goto tr670
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st501
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st93
	st501:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof501
		}
	st_case_501:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr670
		case 11:
			goto tr671
		case 12:
			goto tr493
		case 32:
			goto tr670
		case 34:
			goto tr199
		case 44:
			goto tr184
		case 61:
			goto tr191
		case 92:
			goto st102
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr335
		}
		goto st93
tr641:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st136
tr696:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

	goto st136
tr702:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

	goto st136
tr706:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

	goto st136
tr710:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

	goto st136
	st136:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof136
		}
	st_case_136:
//line plugins/parsers/influx/machine.go:17282
		switch ( m.data)[( m.p)] {
		case 9:
			goto st7
		case 10:
			goto tr59
		case 32:
			goto st7
		case 34:
			goto tr251
		case 44:
			goto st7
		case 61:
			goto st7
		case 92:
			goto tr270
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto tr269
tr269:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st137
	st137:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof137
		}
	st_case_137:
//line plugins/parsers/influx/machine.go:17314
		switch ( m.data)[( m.p)] {
		case 9:
			goto st7
		case 10:
			goto tr59
		case 32:
			goto st7
		case 34:
			goto tr254
		case 44:
			goto st7
		case 61:
			goto tr272
		case 92:
			goto st150
		}
		if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st137
tr272:
//line plugins/parsers/influx/machine.go.rl:76

	key = m.text()

//line plugins/parsers/influx/machine.go.rl:84

	key = m.text()

	goto st138
	st138:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof138
		}
	st_case_138:
//line plugins/parsers/influx/machine.go:17350
		switch ( m.data)[( m.p)] {
		case 9:
			goto st7
		case 10:
			goto tr59
		case 32:
			goto st7
		case 34:
			goto tr258
		case 44:
			goto st7
		case 45:
			goto tr274
		case 46:
			goto tr275
		case 48:
			goto tr276
		case 61:
			goto st7
		case 70:
			goto tr278
		case 84:
			goto tr279
		case 92:
			goto tr179
		case 102:
			goto tr280
		case 116:
			goto tr281
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr277
			}
		case ( m.data)[( m.p)] >= 12:
			goto tr59
		}
		goto tr174
tr274:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st139
	st139:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof139
		}
	st_case_139:
//line plugins/parsers/influx/machine.go:17401
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr181
		case 11:
			goto tr182
		case 12:
			goto tr58
		case 32:
			goto tr181
		case 34:
			goto tr183
		case 44:
			goto tr184
		case 48:
			goto st502
		case 61:
			goto st7
		case 92:
			goto st100
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st507
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr59
		}
		goto st86
	st502:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof502
		}
	st_case_502:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr694
		case 11:
			goto tr695
		case 12:
			goto tr539
		case 32:
			goto tr694
		case 34:
			goto tr183
		case 44:
			goto tr696
		case 46:
			goto st503
		case 61:
			goto st7
		case 69:
			goto st140
		case 92:
			goto st100
		case 101:
			goto st140
		case 105:
			goto st506
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr356
		}
		goto st86
tr275:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st503
	st503:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof503
		}
	st_case_503:
//line plugins/parsers/influx/machine.go:17477
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr694
		case 11:
			goto tr695
		case 12:
			goto tr539
		case 32:
			goto tr694
		case 34:
			goto tr183
		case 44:
			goto tr696
		case 61:
			goto st7
		case 69:
			goto st140
		case 92:
			goto st100
		case 101:
			goto st140
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st503
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr356
		}
		goto st86
	st140:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof140
		}
	st_case_140:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr181
		case 11:
			goto tr182
		case 12:
			goto tr58
		case 32:
			goto tr181
		case 34:
			goto tr284
		case 44:
			goto tr184
		case 61:
			goto st7
		case 92:
			goto st100
		}
		switch {
		case ( m.data)[( m.p)] < 43:
			if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr59
			}
		case ( m.data)[( m.p)] > 45:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st505
			}
		default:
			goto st141
		}
		goto st86
tr284:
//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st504
	st504:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof504
		}
	st_case_504:
//line plugins/parsers/influx/machine.go:17556
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto tr538
		case 13:
			goto tr330
		case 32:
			goto tr487
		case 44:
			goto tr489
		case 61:
			goto tr201
		case 92:
			goto st35
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st373
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr487
		}
		goto st30
	st141:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof141
		}
	st_case_141:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr181
		case 11:
			goto tr182
		case 12:
			goto tr58
		case 32:
			goto tr181
		case 34:
			goto tr183
		case 44:
			goto tr184
		case 61:
			goto st7
		case 92:
			goto st100
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st505
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr59
		}
		goto st86
	st505:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof505
		}
	st_case_505:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr694
		case 11:
			goto tr695
		case 12:
			goto tr539
		case 32:
			goto tr694
		case 34:
			goto tr183
		case 44:
			goto tr696
		case 61:
			goto st7
		case 92:
			goto st100
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st505
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr356
		}
		goto st86
	st506:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof506
		}
	st_case_506:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr700
		case 11:
			goto tr701
		case 12:
			goto tr545
		case 32:
			goto tr700
		case 34:
			goto tr183
		case 44:
			goto tr702
		case 61:
			goto st7
		case 92:
			goto st100
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr362
		}
		goto st86
	st507:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof507
		}
	st_case_507:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr694
		case 11:
			goto tr695
		case 12:
			goto tr539
		case 32:
			goto tr694
		case 34:
			goto tr183
		case 44:
			goto tr696
		case 46:
			goto st503
		case 61:
			goto st7
		case 69:
			goto st140
		case 92:
			goto st100
		case 101:
			goto st140
		case 105:
			goto st506
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st507
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr356
		}
		goto st86
tr276:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st508
	st508:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof508
		}
	st_case_508:
//line plugins/parsers/influx/machine.go:17724
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr694
		case 11:
			goto tr695
		case 12:
			goto tr539
		case 32:
			goto tr694
		case 34:
			goto tr183
		case 44:
			goto tr696
		case 46:
			goto st503
		case 61:
			goto st7
		case 69:
			goto st140
		case 92:
			goto st100
		case 101:
			goto st140
		case 105:
			goto st506
		case 117:
			goto st509
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr356
		}
		goto st86
	st509:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof509
		}
	st_case_509:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr704
		case 11:
			goto tr705
		case 12:
			goto tr549
		case 32:
			goto tr704
		case 34:
			goto tr183
		case 44:
			goto tr706
		case 61:
			goto st7
		case 92:
			goto st100
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr366
		}
		goto st86
tr277:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st510
	st510:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof510
		}
	st_case_510:
//line plugins/parsers/influx/machine.go:17795
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr694
		case 11:
			goto tr695
		case 12:
			goto tr539
		case 32:
			goto tr694
		case 34:
			goto tr183
		case 44:
			goto tr696
		case 46:
			goto st503
		case 61:
			goto st7
		case 69:
			goto st140
		case 92:
			goto st100
		case 101:
			goto st140
		case 105:
			goto st506
		case 117:
			goto st509
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st510
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr356
		}
		goto st86
tr278:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st511
	st511:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof511
		}
	st_case_511:
//line plugins/parsers/influx/machine.go:17844
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr708
		case 11:
			goto tr709
		case 12:
			goto tr553
		case 32:
			goto tr708
		case 34:
			goto tr183
		case 44:
			goto tr710
		case 61:
			goto st7
		case 65:
			goto st142
		case 92:
			goto st100
		case 97:
			goto st145
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr370
		}
		goto st86
	st142:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof142
		}
	st_case_142:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr181
		case 11:
			goto tr182
		case 12:
			goto tr58
		case 32:
			goto tr181
		case 34:
			goto tr183
		case 44:
			goto tr184
		case 61:
			goto st7
		case 76:
			goto st143
		case 92:
			goto st100
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st86
	st143:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof143
		}
	st_case_143:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr181
		case 11:
			goto tr182
		case 12:
			goto tr58
		case 32:
			goto tr181
		case 34:
			goto tr183
		case 44:
			goto tr184
		case 61:
			goto st7
		case 83:
			goto st144
		case 92:
			goto st100
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st86
	st144:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof144
		}
	st_case_144:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr181
		case 11:
			goto tr182
		case 12:
			goto tr58
		case 32:
			goto tr181
		case 34:
			goto tr183
		case 44:
			goto tr184
		case 61:
			goto st7
		case 69:
			goto st512
		case 92:
			goto st100
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st86
	st512:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof512
		}
	st_case_512:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr708
		case 11:
			goto tr709
		case 12:
			goto tr553
		case 32:
			goto tr708
		case 34:
			goto tr183
		case 44:
			goto tr710
		case 61:
			goto st7
		case 92:
			goto st100
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr370
		}
		goto st86
	st145:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof145
		}
	st_case_145:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr181
		case 11:
			goto tr182
		case 12:
			goto tr58
		case 32:
			goto tr181
		case 34:
			goto tr183
		case 44:
			goto tr184
		case 61:
			goto st7
		case 92:
			goto st100
		case 108:
			goto st146
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st86
	st146:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof146
		}
	st_case_146:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr181
		case 11:
			goto tr182
		case 12:
			goto tr58
		case 32:
			goto tr181
		case 34:
			goto tr183
		case 44:
			goto tr184
		case 61:
			goto st7
		case 92:
			goto st100
		case 115:
			goto st147
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st86
	st147:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof147
		}
	st_case_147:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr181
		case 11:
			goto tr182
		case 12:
			goto tr58
		case 32:
			goto tr181
		case 34:
			goto tr183
		case 44:
			goto tr184
		case 61:
			goto st7
		case 92:
			goto st100
		case 101:
			goto st512
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st86
tr279:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st513
	st513:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof513
		}
	st_case_513:
//line plugins/parsers/influx/machine.go:18083
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr708
		case 11:
			goto tr709
		case 12:
			goto tr553
		case 32:
			goto tr708
		case 34:
			goto tr183
		case 44:
			goto tr710
		case 61:
			goto st7
		case 82:
			goto st148
		case 92:
			goto st100
		case 114:
			goto st149
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr370
		}
		goto st86
	st148:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof148
		}
	st_case_148:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr181
		case 11:
			goto tr182
		case 12:
			goto tr58
		case 32:
			goto tr181
		case 34:
			goto tr183
		case 44:
			goto tr184
		case 61:
			goto st7
		case 85:
			goto st144
		case 92:
			goto st100
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st86
	st149:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof149
		}
	st_case_149:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr181
		case 11:
			goto tr182
		case 12:
			goto tr58
		case 32:
			goto tr181
		case 34:
			goto tr183
		case 44:
			goto tr184
		case 61:
			goto st7
		case 92:
			goto st100
		case 117:
			goto st147
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st86
tr280:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st514
	st514:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof514
		}
	st_case_514:
//line plugins/parsers/influx/machine.go:18179
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr708
		case 11:
			goto tr709
		case 12:
			goto tr553
		case 32:
			goto tr708
		case 34:
			goto tr183
		case 44:
			goto tr710
		case 61:
			goto st7
		case 92:
			goto st100
		case 97:
			goto st145
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr370
		}
		goto st86
tr281:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st515
	st515:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof515
		}
	st_case_515:
//line plugins/parsers/influx/machine.go:18215
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr708
		case 11:
			goto tr709
		case 12:
			goto tr553
		case 32:
			goto tr708
		case 34:
			goto tr183
		case 44:
			goto tr710
		case 61:
			goto st7
		case 92:
			goto st100
		case 114:
			goto st149
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr370
		}
		goto st86
tr270:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st150
	st150:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof150
		}
	st_case_150:
//line plugins/parsers/influx/machine.go:18251
		switch ( m.data)[( m.p)] {
		case 34:
			goto st137
		case 92:
			goto st137
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr59
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr59
		}
		goto st83
tr259:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st151
	st151:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof151
		}
	st_case_151:
//line plugins/parsers/influx/machine.go:18278
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr229
		case 11:
			goto tr230
		case 12:
			goto tr58
		case 32:
			goto tr229
		case 34:
			goto tr183
		case 44:
			goto tr231
		case 48:
			goto st516
		case 61:
			goto st7
		case 92:
			goto st123
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st542
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr59
		}
		goto st118
	st516:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof516
		}
	st_case_516:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr715
		case 11:
			goto tr716
		case 12:
			goto tr539
		case 32:
			goto tr715
		case 34:
			goto tr183
		case 44:
			goto tr717
		case 46:
			goto st539
		case 61:
			goto st7
		case 69:
			goto st153
		case 92:
			goto st123
		case 101:
			goto st153
		case 105:
			goto st541
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr356
		}
		goto st118
tr716:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

	goto st517
tr748:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

	goto st517
tr752:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

	goto st517
tr756:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

	goto st517
	st517:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof517
		}
	st_case_517:
//line plugins/parsers/influx/machine.go:18388
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr721
		case 11:
			goto tr722
		case 12:
			goto tr487
		case 32:
			goto tr721
		case 34:
			goto tr195
		case 44:
			goto tr231
		case 45:
			goto tr723
		case 61:
			goto st7
		case 92:
			goto tr235
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr724
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr330
		}
		goto tr233
tr722:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st518
	st518:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof518
		}
	st_case_518:
//line plugins/parsers/influx/machine.go:18433
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr721
		case 11:
			goto tr722
		case 12:
			goto tr487
		case 32:
			goto tr721
		case 34:
			goto tr195
		case 44:
			goto tr231
		case 45:
			goto tr723
		case 61:
			goto tr97
		case 92:
			goto tr235
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr724
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr330
		}
		goto tr233
tr723:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st152
	st152:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof152
		}
	st_case_152:
//line plugins/parsers/influx/machine.go:18474
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr229
		case 11:
			goto tr237
		case 12:
			goto tr58
		case 32:
			goto tr229
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st519
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr201
		}
		goto st120
tr724:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st519
	st519:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof519
		}
	st_case_519:
//line plugins/parsers/influx/machine.go:18513
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr725
		case 11:
			goto tr726
		case 12:
			goto tr493
		case 32:
			goto tr725
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st521
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st120
tr729:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st520
tr726:
//line plugins/parsers/influx/machine.go.rl:80

	m.handler.AddTag(key, m.text())

//line plugins/parsers/influx/machine.go.rl:108

	m.handler.SetTimestamp(m.text())

	goto st520
	st520:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof520
		}
	st_case_520:
//line plugins/parsers/influx/machine.go:18566
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr728
		case 11:
			goto tr729
		case 12:
			goto tr496
		case 32:
			goto tr728
		case 34:
			goto tr195
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto tr235
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr330
		}
		goto tr233
	st521:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof521
		}
	st_case_521:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr725
		case 11:
			goto tr726
		case 12:
			goto tr493
		case 32:
			goto tr725
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st522
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st120
	st522:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof522
		}
	st_case_522:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr725
		case 11:
			goto tr726
		case 12:
			goto tr493
		case 32:
			goto tr725
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st523
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st120
	st523:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof523
		}
	st_case_523:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr725
		case 11:
			goto tr726
		case 12:
			goto tr493
		case 32:
			goto tr725
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st524
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st120
	st524:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof524
		}
	st_case_524:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr725
		case 11:
			goto tr726
		case 12:
			goto tr493
		case 32:
			goto tr725
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st525
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st120
	st525:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof525
		}
	st_case_525:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr725
		case 11:
			goto tr726
		case 12:
			goto tr493
		case 32:
			goto tr725
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st526
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st120
	st526:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof526
		}
	st_case_526:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr725
		case 11:
			goto tr726
		case 12:
			goto tr493
		case 32:
			goto tr725
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st527
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st120
	st527:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof527
		}
	st_case_527:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr725
		case 11:
			goto tr726
		case 12:
			goto tr493
		case 32:
			goto tr725
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st528
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st120
	st528:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof528
		}
	st_case_528:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr725
		case 11:
			goto tr726
		case 12:
			goto tr493
		case 32:
			goto tr725
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st529
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st120
	st529:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof529
		}
	st_case_529:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr725
		case 11:
			goto tr726
		case 12:
			goto tr493
		case 32:
			goto tr725
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st530
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st120
	st530:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof530
		}
	st_case_530:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr725
		case 11:
			goto tr726
		case 12:
			goto tr493
		case 32:
			goto tr725
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st531
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st120
	st531:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof531
		}
	st_case_531:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr725
		case 11:
			goto tr726
		case 12:
			goto tr493
		case 32:
			goto tr725
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st532
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st120
	st532:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof532
		}
	st_case_532:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr725
		case 11:
			goto tr726
		case 12:
			goto tr493
		case 32:
			goto tr725
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st533
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st120
	st533:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof533
		}
	st_case_533:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr725
		case 11:
			goto tr726
		case 12:
			goto tr493
		case 32:
			goto tr725
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st534
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st120
	st534:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof534
		}
	st_case_534:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr725
		case 11:
			goto tr726
		case 12:
			goto tr493
		case 32:
			goto tr725
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st535
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st120
	st535:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof535
		}
	st_case_535:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr725
		case 11:
			goto tr726
		case 12:
			goto tr493
		case 32:
			goto tr725
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st536
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st120
	st536:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof536
		}
	st_case_536:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr725
		case 11:
			goto tr726
		case 12:
			goto tr493
		case 32:
			goto tr725
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st537
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st120
	st537:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof537
		}
	st_case_537:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr725
		case 11:
			goto tr726
		case 12:
			goto tr493
		case 32:
			goto tr725
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st538
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr335
		}
		goto st120
	st538:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof538
		}
	st_case_538:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr725
		case 11:
			goto tr726
		case 12:
			goto tr493
		case 32:
			goto tr725
		case 34:
			goto tr199
		case 44:
			goto tr231
		case 61:
			goto tr97
		case 92:
			goto st122
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr335
		}
		goto st120
tr260:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st539
	st539:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof539
		}
	st_case_539:
//line plugins/parsers/influx/machine.go:19171
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr715
		case 11:
			goto tr716
		case 12:
			goto tr539
		case 32:
			goto tr715
		case 34:
			goto tr183
		case 44:
			goto tr717
		case 61:
			goto st7
		case 69:
			goto st153
		case 92:
			goto st123
		case 101:
			goto st153
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st539
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr356
		}
		goto st118
	st153:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof153
		}
	st_case_153:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr229
		case 11:
			goto tr230
		case 12:
			goto tr58
		case 32:
			goto tr229
		case 34:
			goto tr284
		case 44:
			goto tr231
		case 61:
			goto st7
		case 92:
			goto st123
		}
		switch {
		case ( m.data)[( m.p)] < 43:
			if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr59
			}
		case ( m.data)[( m.p)] > 45:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st540
			}
		default:
			goto st154
		}
		goto st118
	st154:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof154
		}
	st_case_154:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr229
		case 11:
			goto tr230
		case 12:
			goto tr58
		case 32:
			goto tr229
		case 34:
			goto tr183
		case 44:
			goto tr231
		case 61:
			goto st7
		case 92:
			goto st123
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st540
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr59
		}
		goto st118
	st540:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof540
		}
	st_case_540:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr715
		case 11:
			goto tr716
		case 12:
			goto tr539
		case 32:
			goto tr715
		case 34:
			goto tr183
		case 44:
			goto tr717
		case 61:
			goto st7
		case 92:
			goto st123
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st540
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr356
		}
		goto st118
	st541:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof541
		}
	st_case_541:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr747
		case 11:
			goto tr748
		case 12:
			goto tr545
		case 32:
			goto tr747
		case 34:
			goto tr183
		case 44:
			goto tr749
		case 61:
			goto st7
		case 92:
			goto st123
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr362
		}
		goto st118
	st542:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof542
		}
	st_case_542:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr715
		case 11:
			goto tr716
		case 12:
			goto tr539
		case 32:
			goto tr715
		case 34:
			goto tr183
		case 44:
			goto tr717
		case 46:
			goto st539
		case 61:
			goto st7
		case 69:
			goto st153
		case 92:
			goto st123
		case 101:
			goto st153
		case 105:
			goto st541
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st542
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr356
		}
		goto st118
tr261:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st543
	st543:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof543
		}
	st_case_543:
//line plugins/parsers/influx/machine.go:19381
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr715
		case 11:
			goto tr716
		case 12:
			goto tr539
		case 32:
			goto tr715
		case 34:
			goto tr183
		case 44:
			goto tr717
		case 46:
			goto st539
		case 61:
			goto st7
		case 69:
			goto st153
		case 92:
			goto st123
		case 101:
			goto st153
		case 105:
			goto st541
		case 117:
			goto st544
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr356
		}
		goto st118
	st544:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof544
		}
	st_case_544:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr751
		case 11:
			goto tr752
		case 12:
			goto tr549
		case 32:
			goto tr751
		case 34:
			goto tr183
		case 44:
			goto tr753
		case 61:
			goto st7
		case 92:
			goto st123
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr366
		}
		goto st118
tr262:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st545
	st545:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof545
		}
	st_case_545:
//line plugins/parsers/influx/machine.go:19452
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr715
		case 11:
			goto tr716
		case 12:
			goto tr539
		case 32:
			goto tr715
		case 34:
			goto tr183
		case 44:
			goto tr717
		case 46:
			goto st539
		case 61:
			goto st7
		case 69:
			goto st153
		case 92:
			goto st123
		case 101:
			goto st153
		case 105:
			goto st541
		case 117:
			goto st544
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st545
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr356
		}
		goto st118
tr263:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st546
	st546:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof546
		}
	st_case_546:
//line plugins/parsers/influx/machine.go:19501
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr755
		case 11:
			goto tr756
		case 12:
			goto tr553
		case 32:
			goto tr755
		case 34:
			goto tr183
		case 44:
			goto tr757
		case 61:
			goto st7
		case 65:
			goto st155
		case 92:
			goto st123
		case 97:
			goto st158
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr370
		}
		goto st118
	st155:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof155
		}
	st_case_155:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr229
		case 11:
			goto tr230
		case 12:
			goto tr58
		case 32:
			goto tr229
		case 34:
			goto tr183
		case 44:
			goto tr231
		case 61:
			goto st7
		case 76:
			goto st156
		case 92:
			goto st123
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st118
	st156:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof156
		}
	st_case_156:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr229
		case 11:
			goto tr230
		case 12:
			goto tr58
		case 32:
			goto tr229
		case 34:
			goto tr183
		case 44:
			goto tr231
		case 61:
			goto st7
		case 83:
			goto st157
		case 92:
			goto st123
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st118
	st157:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof157
		}
	st_case_157:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr229
		case 11:
			goto tr230
		case 12:
			goto tr58
		case 32:
			goto tr229
		case 34:
			goto tr183
		case 44:
			goto tr231
		case 61:
			goto st7
		case 69:
			goto st547
		case 92:
			goto st123
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st118
	st547:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof547
		}
	st_case_547:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr755
		case 11:
			goto tr756
		case 12:
			goto tr553
		case 32:
			goto tr755
		case 34:
			goto tr183
		case 44:
			goto tr757
		case 61:
			goto st7
		case 92:
			goto st123
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr370
		}
		goto st118
	st158:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof158
		}
	st_case_158:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr229
		case 11:
			goto tr230
		case 12:
			goto tr58
		case 32:
			goto tr229
		case 34:
			goto tr183
		case 44:
			goto tr231
		case 61:
			goto st7
		case 92:
			goto st123
		case 108:
			goto st159
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st118
	st159:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof159
		}
	st_case_159:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr229
		case 11:
			goto tr230
		case 12:
			goto tr58
		case 32:
			goto tr229
		case 34:
			goto tr183
		case 44:
			goto tr231
		case 61:
			goto st7
		case 92:
			goto st123
		case 115:
			goto st160
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st118
	st160:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof160
		}
	st_case_160:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr229
		case 11:
			goto tr230
		case 12:
			goto tr58
		case 32:
			goto tr229
		case 34:
			goto tr183
		case 44:
			goto tr231
		case 61:
			goto st7
		case 92:
			goto st123
		case 101:
			goto st547
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st118
tr264:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st548
	st548:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof548
		}
	st_case_548:
//line plugins/parsers/influx/machine.go:19740
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr755
		case 11:
			goto tr756
		case 12:
			goto tr553
		case 32:
			goto tr755
		case 34:
			goto tr183
		case 44:
			goto tr757
		case 61:
			goto st7
		case 82:
			goto st161
		case 92:
			goto st123
		case 114:
			goto st162
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr370
		}
		goto st118
	st161:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof161
		}
	st_case_161:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr229
		case 11:
			goto tr230
		case 12:
			goto tr58
		case 32:
			goto tr229
		case 34:
			goto tr183
		case 44:
			goto tr231
		case 61:
			goto st7
		case 85:
			goto st157
		case 92:
			goto st123
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st118
	st162:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof162
		}
	st_case_162:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr229
		case 11:
			goto tr230
		case 12:
			goto tr58
		case 32:
			goto tr229
		case 34:
			goto tr183
		case 44:
			goto tr231
		case 61:
			goto st7
		case 92:
			goto st123
		case 117:
			goto st160
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr59
		}
		goto st118
tr265:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st549
	st549:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof549
		}
	st_case_549:
//line plugins/parsers/influx/machine.go:19836
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr755
		case 11:
			goto tr756
		case 12:
			goto tr553
		case 32:
			goto tr755
		case 34:
			goto tr183
		case 44:
			goto tr757
		case 61:
			goto st7
		case 92:
			goto st123
		case 97:
			goto st158
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr370
		}
		goto st118
tr266:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st550
	st550:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof550
		}
	st_case_550:
//line plugins/parsers/influx/machine.go:19872
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr755
		case 11:
			goto tr756
		case 12:
			goto tr553
		case 32:
			goto tr755
		case 34:
			goto tr183
		case 44:
			goto tr757
		case 61:
			goto st7
		case 92:
			goto st123
		case 114:
			goto st162
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr370
		}
		goto st118
tr252:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st163
	st163:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof163
		}
	st_case_163:
//line plugins/parsers/influx/machine.go:19908
		switch ( m.data)[( m.p)] {
		case 34:
			goto st131
		case 92:
			goto st131
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr59
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr59
		}
		goto st83
tr83:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st164
	st164:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof164
		}
	st_case_164:
//line plugins/parsers/influx/machine.go:19935
		switch ( m.data)[( m.p)] {
		case 34:
			goto st39
		case 92:
			goto st39
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr5
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr5
		}
		goto st2
tr240:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st165
	st165:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof165
		}
	st_case_165:
//line plugins/parsers/influx/machine.go:19962
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr85
		case 11:
			goto tr86
		case 12:
			goto tr4
		case 32:
			goto tr85
		case 34:
			goto tr87
		case 44:
			goto tr88
		case 48:
			goto st551
		case 92:
			goto st164
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st557
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr5
		}
		goto st39
	st551:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof551
		}
	st_case_551:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr762
		case 11:
			goto tr763
		case 12:
			goto tr764
		case 32:
			goto tr762
		case 34:
			goto tr87
		case 44:
			goto tr765
		case 46:
			goto st552
		case 69:
			goto st166
		case 92:
			goto st164
		case 101:
			goto st166
		case 105:
			goto st556
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr356
		}
		goto st39
tr241:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st552
	st552:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof552
		}
	st_case_552:
//line plugins/parsers/influx/machine.go:20034
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr762
		case 11:
			goto tr763
		case 12:
			goto tr764
		case 32:
			goto tr762
		case 34:
			goto tr87
		case 44:
			goto tr765
		case 69:
			goto st166
		case 92:
			goto st164
		case 101:
			goto st166
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st552
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr356
		}
		goto st39
	st166:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof166
		}
	st_case_166:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr85
		case 11:
			goto tr86
		case 12:
			goto tr4
		case 32:
			goto tr85
		case 34:
			goto tr304
		case 44:
			goto tr88
		case 92:
			goto st164
		}
		switch {
		case ( m.data)[( m.p)] < 43:
			if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr5
			}
		case ( m.data)[( m.p)] > 45:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st555
			}
		default:
			goto st167
		}
		goto st39
tr304:
//line plugins/parsers/influx/machine.go.rl:104

	m.handler.AddString(key, m.text())

	goto st553
	st553:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof553
		}
	st_case_553:
//line plugins/parsers/influx/machine.go:20109
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr330
		case 11:
			goto tr615
		case 13:
			goto tr330
		case 32:
			goto tr455
		case 44:
			goto tr457
		case 92:
			goto st129
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st554
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr455
		}
		goto st2
	st554:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof554
		}
	st_case_554:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 11:
			goto tr769
		case 13:
			goto tr356
		case 32:
			goto tr764
		case 44:
			goto tr770
		case 92:
			goto st129
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st554
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr764
		}
		goto st2
	st167:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof167
		}
	st_case_167:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr85
		case 11:
			goto tr86
		case 12:
			goto tr4
		case 32:
			goto tr85
		case 34:
			goto tr87
		case 44:
			goto tr88
		case 92:
			goto st164
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st555
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr5
		}
		goto st39
	st555:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof555
		}
	st_case_555:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr762
		case 11:
			goto tr763
		case 12:
			goto tr764
		case 32:
			goto tr762
		case 34:
			goto tr87
		case 44:
			goto tr765
		case 92:
			goto st164
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st555
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr356
		}
		goto st39
	st556:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof556
		}
	st_case_556:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr771
		case 11:
			goto tr772
		case 12:
			goto tr773
		case 32:
			goto tr771
		case 34:
			goto tr87
		case 44:
			goto tr774
		case 92:
			goto st164
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr362
		}
		goto st39
	st557:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof557
		}
	st_case_557:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr762
		case 11:
			goto tr763
		case 12:
			goto tr764
		case 32:
			goto tr762
		case 34:
			goto tr87
		case 44:
			goto tr765
		case 46:
			goto st552
		case 69:
			goto st166
		case 92:
			goto st164
		case 101:
			goto st166
		case 105:
			goto st556
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st557
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr356
		}
		goto st39
tr242:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st558
	st558:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof558
		}
	st_case_558:
//line plugins/parsers/influx/machine.go:20295
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr762
		case 11:
			goto tr763
		case 12:
			goto tr764
		case 32:
			goto tr762
		case 34:
			goto tr87
		case 44:
			goto tr765
		case 46:
			goto st552
		case 69:
			goto st166
		case 92:
			goto st164
		case 101:
			goto st166
		case 105:
			goto st556
		case 117:
			goto st559
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr356
		}
		goto st39
	st559:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof559
		}
	st_case_559:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr776
		case 11:
			goto tr777
		case 12:
			goto tr778
		case 32:
			goto tr776
		case 34:
			goto tr87
		case 44:
			goto tr779
		case 92:
			goto st164
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr366
		}
		goto st39
tr243:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st560
	st560:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof560
		}
	st_case_560:
//line plugins/parsers/influx/machine.go:20362
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr762
		case 11:
			goto tr763
		case 12:
			goto tr764
		case 32:
			goto tr762
		case 34:
			goto tr87
		case 44:
			goto tr765
		case 46:
			goto st552
		case 69:
			goto st166
		case 92:
			goto st164
		case 101:
			goto st166
		case 105:
			goto st556
		case 117:
			goto st559
		}
		switch {
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st560
			}
		case ( m.data)[( m.p)] >= 10:
			goto tr356
		}
		goto st39
tr244:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st561
	st561:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof561
		}
	st_case_561:
//line plugins/parsers/influx/machine.go:20409
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr781
		case 11:
			goto tr782
		case 12:
			goto tr783
		case 32:
			goto tr781
		case 34:
			goto tr87
		case 44:
			goto tr784
		case 65:
			goto st168
		case 92:
			goto st164
		case 97:
			goto st171
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr370
		}
		goto st39
	st168:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof168
		}
	st_case_168:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr85
		case 11:
			goto tr86
		case 12:
			goto tr4
		case 32:
			goto tr85
		case 34:
			goto tr87
		case 44:
			goto tr88
		case 76:
			goto st169
		case 92:
			goto st164
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st39
	st169:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof169
		}
	st_case_169:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr85
		case 11:
			goto tr86
		case 12:
			goto tr4
		case 32:
			goto tr85
		case 34:
			goto tr87
		case 44:
			goto tr88
		case 83:
			goto st170
		case 92:
			goto st164
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st39
	st170:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof170
		}
	st_case_170:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr85
		case 11:
			goto tr86
		case 12:
			goto tr4
		case 32:
			goto tr85
		case 34:
			goto tr87
		case 44:
			goto tr88
		case 69:
			goto st562
		case 92:
			goto st164
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st39
	st562:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof562
		}
	st_case_562:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr781
		case 11:
			goto tr782
		case 12:
			goto tr783
		case 32:
			goto tr781
		case 34:
			goto tr87
		case 44:
			goto tr784
		case 92:
			goto st164
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr370
		}
		goto st39
	st171:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof171
		}
	st_case_171:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr85
		case 11:
			goto tr86
		case 12:
			goto tr4
		case 32:
			goto tr85
		case 34:
			goto tr87
		case 44:
			goto tr88
		case 92:
			goto st164
		case 108:
			goto st172
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st39
	st172:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof172
		}
	st_case_172:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr85
		case 11:
			goto tr86
		case 12:
			goto tr4
		case 32:
			goto tr85
		case 34:
			goto tr87
		case 44:
			goto tr88
		case 92:
			goto st164
		case 115:
			goto st173
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st39
	st173:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof173
		}
	st_case_173:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr85
		case 11:
			goto tr86
		case 12:
			goto tr4
		case 32:
			goto tr85
		case 34:
			goto tr87
		case 44:
			goto tr88
		case 92:
			goto st164
		case 101:
			goto st562
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st39
tr245:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st563
	st563:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof563
		}
	st_case_563:
//line plugins/parsers/influx/machine.go:20632
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr781
		case 11:
			goto tr782
		case 12:
			goto tr783
		case 32:
			goto tr781
		case 34:
			goto tr87
		case 44:
			goto tr784
		case 82:
			goto st174
		case 92:
			goto st164
		case 114:
			goto st175
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr370
		}
		goto st39
	st174:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof174
		}
	st_case_174:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr85
		case 11:
			goto tr86
		case 12:
			goto tr4
		case 32:
			goto tr85
		case 34:
			goto tr87
		case 44:
			goto tr88
		case 85:
			goto st170
		case 92:
			goto st164
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st39
	st175:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof175
		}
	st_case_175:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr85
		case 11:
			goto tr86
		case 12:
			goto tr4
		case 32:
			goto tr85
		case 34:
			goto tr87
		case 44:
			goto tr88
		case 92:
			goto st164
		case 117:
			goto st173
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr5
		}
		goto st39
tr246:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st564
	st564:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof564
		}
	st_case_564:
//line plugins/parsers/influx/machine.go:20722
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr781
		case 11:
			goto tr782
		case 12:
			goto tr783
		case 32:
			goto tr781
		case 34:
			goto tr87
		case 44:
			goto tr784
		case 92:
			goto st164
		case 97:
			goto st171
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr370
		}
		goto st39
tr247:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st565
	st565:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof565
		}
	st_case_565:
//line plugins/parsers/influx/machine.go:20756
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr781
		case 11:
			goto tr782
		case 12:
			goto tr783
		case 32:
			goto tr781
		case 34:
			goto tr87
		case 44:
			goto tr784
		case 92:
			goto st164
		case 114:
			goto st175
		}
		if 10 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto tr370
		}
		goto st39
tr70:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st176
	st176:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof176
		}
	st_case_176:
//line plugins/parsers/influx/machine.go:20790
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 11:
			goto tr6
		case 13:
			goto tr5
		case 32:
			goto tr4
		case 44:
			goto tr7
		case 48:
			goto st566
		case 92:
			goto st129
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st569
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr4
		}
		goto st2
	st566:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof566
		}
	st_case_566:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 11:
			goto tr769
		case 13:
			goto tr356
		case 32:
			goto tr764
		case 44:
			goto tr770
		case 46:
			goto st567
		case 69:
			goto st177
		case 92:
			goto st129
		case 101:
			goto st177
		case 105:
			goto st568
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr764
		}
		goto st2
tr71:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st567
	st567:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof567
		}
	st_case_567:
//line plugins/parsers/influx/machine.go:20858
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 11:
			goto tr769
		case 13:
			goto tr356
		case 32:
			goto tr764
		case 44:
			goto tr770
		case 69:
			goto st177
		case 92:
			goto st129
		case 101:
			goto st177
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st567
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr764
		}
		goto st2
	st177:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof177
		}
	st_case_177:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 11:
			goto tr6
		case 13:
			goto tr5
		case 32:
			goto tr4
		case 34:
			goto st178
		case 44:
			goto tr7
		case 92:
			goto st129
		}
		switch {
		case ( m.data)[( m.p)] < 43:
			if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
				goto tr4
			}
		case ( m.data)[( m.p)] > 45:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st554
			}
		default:
			goto st178
		}
		goto st2
	st178:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof178
		}
	st_case_178:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 11:
			goto tr6
		case 13:
			goto tr5
		case 32:
			goto tr4
		case 44:
			goto tr7
		case 92:
			goto st129
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st554
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr4
		}
		goto st2
	st568:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof568
		}
	st_case_568:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr362
		case 11:
			goto tr792
		case 13:
			goto tr362
		case 32:
			goto tr773
		case 44:
			goto tr793
		case 92:
			goto st129
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr773
		}
		goto st2
	st569:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof569
		}
	st_case_569:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 11:
			goto tr769
		case 13:
			goto tr356
		case 32:
			goto tr764
		case 44:
			goto tr770
		case 46:
			goto st567
		case 69:
			goto st177
		case 92:
			goto st129
		case 101:
			goto st177
		case 105:
			goto st568
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st569
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr764
		}
		goto st2
tr72:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st570
	st570:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof570
		}
	st_case_570:
//line plugins/parsers/influx/machine.go:21018
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 11:
			goto tr769
		case 13:
			goto tr356
		case 32:
			goto tr764
		case 44:
			goto tr770
		case 46:
			goto st567
		case 69:
			goto st177
		case 92:
			goto st129
		case 101:
			goto st177
		case 105:
			goto st568
		case 117:
			goto st571
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr764
		}
		goto st2
	st571:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof571
		}
	st_case_571:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr366
		case 11:
			goto tr795
		case 13:
			goto tr366
		case 32:
			goto tr778
		case 44:
			goto tr796
		case 92:
			goto st129
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr778
		}
		goto st2
tr73:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st572
	st572:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof572
		}
	st_case_572:
//line plugins/parsers/influx/machine.go:21081
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr356
		case 11:
			goto tr769
		case 13:
			goto tr356
		case 32:
			goto tr764
		case 44:
			goto tr770
		case 46:
			goto st567
		case 69:
			goto st177
		case 92:
			goto st129
		case 101:
			goto st177
		case 105:
			goto st568
		case 117:
			goto st571
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st572
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr764
		}
		goto st2
tr74:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st573
	st573:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof573
		}
	st_case_573:
//line plugins/parsers/influx/machine.go:21126
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 11:
			goto tr798
		case 13:
			goto tr370
		case 32:
			goto tr783
		case 44:
			goto tr799
		case 65:
			goto st179
		case 92:
			goto st129
		case 97:
			goto st182
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr783
		}
		goto st2
	st179:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof179
		}
	st_case_179:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 11:
			goto tr6
		case 13:
			goto tr5
		case 32:
			goto tr4
		case 44:
			goto tr7
		case 76:
			goto st180
		case 92:
			goto st129
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr4
		}
		goto st2
	st180:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof180
		}
	st_case_180:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 11:
			goto tr6
		case 13:
			goto tr5
		case 32:
			goto tr4
		case 44:
			goto tr7
		case 83:
			goto st181
		case 92:
			goto st129
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr4
		}
		goto st2
	st181:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof181
		}
	st_case_181:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 11:
			goto tr6
		case 13:
			goto tr5
		case 32:
			goto tr4
		case 44:
			goto tr7
		case 69:
			goto st574
		case 92:
			goto st129
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr4
		}
		goto st2
	st574:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof574
		}
	st_case_574:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 11:
			goto tr798
		case 13:
			goto tr370
		case 32:
			goto tr783
		case 44:
			goto tr799
		case 92:
			goto st129
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr783
		}
		goto st2
	st182:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof182
		}
	st_case_182:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 11:
			goto tr6
		case 13:
			goto tr5
		case 32:
			goto tr4
		case 44:
			goto tr7
		case 92:
			goto st129
		case 108:
			goto st183
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr4
		}
		goto st2
	st183:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof183
		}
	st_case_183:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 11:
			goto tr6
		case 13:
			goto tr5
		case 32:
			goto tr4
		case 44:
			goto tr7
		case 92:
			goto st129
		case 115:
			goto st184
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr4
		}
		goto st2
	st184:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof184
		}
	st_case_184:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 11:
			goto tr6
		case 13:
			goto tr5
		case 32:
			goto tr4
		case 44:
			goto tr7
		case 92:
			goto st129
		case 101:
			goto st574
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr4
		}
		goto st2
tr75:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st575
	st575:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof575
		}
	st_case_575:
//line plugins/parsers/influx/machine.go:21333
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 11:
			goto tr798
		case 13:
			goto tr370
		case 32:
			goto tr783
		case 44:
			goto tr799
		case 82:
			goto st185
		case 92:
			goto st129
		case 114:
			goto st186
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr783
		}
		goto st2
	st185:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof185
		}
	st_case_185:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 11:
			goto tr6
		case 13:
			goto tr5
		case 32:
			goto tr4
		case 44:
			goto tr7
		case 85:
			goto st181
		case 92:
			goto st129
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr4
		}
		goto st2
	st186:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof186
		}
	st_case_186:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr5
		case 11:
			goto tr6
		case 13:
			goto tr5
		case 32:
			goto tr4
		case 44:
			goto tr7
		case 92:
			goto st129
		case 117:
			goto st184
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr4
		}
		goto st2
tr76:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st576
	st576:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof576
		}
	st_case_576:
//line plugins/parsers/influx/machine.go:21417
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 11:
			goto tr798
		case 13:
			goto tr370
		case 32:
			goto tr783
		case 44:
			goto tr799
		case 92:
			goto st129
		case 97:
			goto st182
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr783
		}
		goto st2
tr77:
//line plugins/parsers/influx/machine.go.rl:18

	m.pb = m.p

	goto st577
	st577:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof577
		}
	st_case_577:
//line plugins/parsers/influx/machine.go:21449
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr370
		case 11:
			goto tr798
		case 13:
			goto tr370
		case 32:
			goto tr783
		case 44:
			goto tr799
		case 92:
			goto st129
		case 114:
			goto st186
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr783
		}
		goto st2
	st187:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof187
		}
	st_case_187:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr322
		case 13:
			goto tr322
		}
		goto st187
tr322:
//line plugins/parsers/influx/machine.go.rl:68

	{goto st188 }

	goto st578
	st578:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof578
		}
	st_case_578:
//line plugins/parsers/influx/machine.go:21493
		goto st0
	st188:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof188
		}
	st_case_188:
		switch ( m.data)[( m.p)] {
		case 11:
			goto tr325
		case 32:
			goto st188
		case 35:
			goto st189
		case 44:
			goto st0
		case 92:
			goto st190
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto st188
		}
		goto tr323
tr323:
//line plugins/parsers/influx/machine.go.rl:63

	( m.p)--

	{goto st1 }

	goto st579
	st579:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof579
		}
	st_case_579:
//line plugins/parsers/influx/machine.go:21529
		goto st0
tr325:
//line plugins/parsers/influx/machine.go.rl:63

	( m.p)--

	{goto st1 }

	goto st580
	st580:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof580
		}
	st_case_580:
//line plugins/parsers/influx/machine.go:21544
		switch ( m.data)[( m.p)] {
		case 11:
			goto tr325
		case 32:
			goto st188
		case 35:
			goto st189
		case 44:
			goto st0
		case 92:
			goto st190
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
			goto st188
		}
		goto tr323
	st189:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof189
		}
	st_case_189:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st188
		case 13:
			goto st188
		}
		goto st189
	st190:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof190
		}
	st_case_190:
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto st0
			}
		case ( m.data)[( m.p)] >= 9:
			goto st0
		}
		goto tr323
	st_out:
	_test_eof1:  m.cs = 1; goto _test_eof
	_test_eof2:  m.cs = 2; goto _test_eof
	_test_eof3:  m.cs = 3; goto _test_eof
	_test_eof4:  m.cs = 4; goto _test_eof
	_test_eof5:  m.cs = 5; goto _test_eof
	_test_eof6:  m.cs = 6; goto _test_eof
	_test_eof7:  m.cs = 7; goto _test_eof
	_test_eof191:  m.cs = 191; goto _test_eof
	_test_eof192:  m.cs = 192; goto _test_eof
	_test_eof193:  m.cs = 193; goto _test_eof
	_test_eof8:  m.cs = 8; goto _test_eof
	_test_eof194:  m.cs = 194; goto _test_eof
	_test_eof195:  m.cs = 195; goto _test_eof
	_test_eof196:  m.cs = 196; goto _test_eof
	_test_eof197:  m.cs = 197; goto _test_eof
	_test_eof198:  m.cs = 198; goto _test_eof
	_test_eof199:  m.cs = 199; goto _test_eof
	_test_eof200:  m.cs = 200; goto _test_eof
	_test_eof201:  m.cs = 201; goto _test_eof
	_test_eof202:  m.cs = 202; goto _test_eof
	_test_eof203:  m.cs = 203; goto _test_eof
	_test_eof204:  m.cs = 204; goto _test_eof
	_test_eof205:  m.cs = 205; goto _test_eof
	_test_eof206:  m.cs = 206; goto _test_eof
	_test_eof207:  m.cs = 207; goto _test_eof
	_test_eof208:  m.cs = 208; goto _test_eof
	_test_eof209:  m.cs = 209; goto _test_eof
	_test_eof210:  m.cs = 210; goto _test_eof
	_test_eof211:  m.cs = 211; goto _test_eof
	_test_eof212:  m.cs = 212; goto _test_eof
	_test_eof213:  m.cs = 213; goto _test_eof
	_test_eof9:  m.cs = 9; goto _test_eof
	_test_eof10:  m.cs = 10; goto _test_eof
	_test_eof11:  m.cs = 11; goto _test_eof
	_test_eof12:  m.cs = 12; goto _test_eof
	_test_eof214:  m.cs = 214; goto _test_eof
	_test_eof215:  m.cs = 215; goto _test_eof
	_test_eof13:  m.cs = 13; goto _test_eof
	_test_eof14:  m.cs = 14; goto _test_eof
	_test_eof216:  m.cs = 216; goto _test_eof
	_test_eof217:  m.cs = 217; goto _test_eof
	_test_eof218:  m.cs = 218; goto _test_eof
	_test_eof219:  m.cs = 219; goto _test_eof
	_test_eof220:  m.cs = 220; goto _test_eof
	_test_eof221:  m.cs = 221; goto _test_eof
	_test_eof222:  m.cs = 222; goto _test_eof
	_test_eof15:  m.cs = 15; goto _test_eof
	_test_eof16:  m.cs = 16; goto _test_eof
	_test_eof17:  m.cs = 17; goto _test_eof
	_test_eof223:  m.cs = 223; goto _test_eof
	_test_eof18:  m.cs = 18; goto _test_eof
	_test_eof19:  m.cs = 19; goto _test_eof
	_test_eof20:  m.cs = 20; goto _test_eof
	_test_eof224:  m.cs = 224; goto _test_eof
	_test_eof21:  m.cs = 21; goto _test_eof
	_test_eof22:  m.cs = 22; goto _test_eof
	_test_eof225:  m.cs = 225; goto _test_eof
	_test_eof226:  m.cs = 226; goto _test_eof
	_test_eof23:  m.cs = 23; goto _test_eof
	_test_eof24:  m.cs = 24; goto _test_eof
	_test_eof25:  m.cs = 25; goto _test_eof
	_test_eof26:  m.cs = 26; goto _test_eof
	_test_eof27:  m.cs = 27; goto _test_eof
	_test_eof28:  m.cs = 28; goto _test_eof
	_test_eof29:  m.cs = 29; goto _test_eof
	_test_eof30:  m.cs = 30; goto _test_eof
	_test_eof31:  m.cs = 31; goto _test_eof
	_test_eof32:  m.cs = 32; goto _test_eof
	_test_eof33:  m.cs = 33; goto _test_eof
	_test_eof34:  m.cs = 34; goto _test_eof
	_test_eof35:  m.cs = 35; goto _test_eof
	_test_eof36:  m.cs = 36; goto _test_eof
	_test_eof37:  m.cs = 37; goto _test_eof
	_test_eof38:  m.cs = 38; goto _test_eof
	_test_eof39:  m.cs = 39; goto _test_eof
	_test_eof40:  m.cs = 40; goto _test_eof
	_test_eof41:  m.cs = 41; goto _test_eof
	_test_eof227:  m.cs = 227; goto _test_eof
	_test_eof228:  m.cs = 228; goto _test_eof
	_test_eof42:  m.cs = 42; goto _test_eof
	_test_eof229:  m.cs = 229; goto _test_eof
	_test_eof230:  m.cs = 230; goto _test_eof
	_test_eof231:  m.cs = 231; goto _test_eof
	_test_eof232:  m.cs = 232; goto _test_eof
	_test_eof233:  m.cs = 233; goto _test_eof
	_test_eof234:  m.cs = 234; goto _test_eof
	_test_eof235:  m.cs = 235; goto _test_eof
	_test_eof236:  m.cs = 236; goto _test_eof
	_test_eof237:  m.cs = 237; goto _test_eof
	_test_eof238:  m.cs = 238; goto _test_eof
	_test_eof239:  m.cs = 239; goto _test_eof
	_test_eof240:  m.cs = 240; goto _test_eof
	_test_eof241:  m.cs = 241; goto _test_eof
	_test_eof242:  m.cs = 242; goto _test_eof
	_test_eof243:  m.cs = 243; goto _test_eof
	_test_eof244:  m.cs = 244; goto _test_eof
	_test_eof245:  m.cs = 245; goto _test_eof
	_test_eof246:  m.cs = 246; goto _test_eof
	_test_eof247:  m.cs = 247; goto _test_eof
	_test_eof248:  m.cs = 248; goto _test_eof
	_test_eof43:  m.cs = 43; goto _test_eof
	_test_eof249:  m.cs = 249; goto _test_eof
	_test_eof250:  m.cs = 250; goto _test_eof
	_test_eof44:  m.cs = 44; goto _test_eof
	_test_eof251:  m.cs = 251; goto _test_eof
	_test_eof252:  m.cs = 252; goto _test_eof
	_test_eof253:  m.cs = 253; goto _test_eof
	_test_eof254:  m.cs = 254; goto _test_eof
	_test_eof255:  m.cs = 255; goto _test_eof
	_test_eof256:  m.cs = 256; goto _test_eof
	_test_eof257:  m.cs = 257; goto _test_eof
	_test_eof258:  m.cs = 258; goto _test_eof
	_test_eof259:  m.cs = 259; goto _test_eof
	_test_eof260:  m.cs = 260; goto _test_eof
	_test_eof261:  m.cs = 261; goto _test_eof
	_test_eof262:  m.cs = 262; goto _test_eof
	_test_eof263:  m.cs = 263; goto _test_eof
	_test_eof264:  m.cs = 264; goto _test_eof
	_test_eof265:  m.cs = 265; goto _test_eof
	_test_eof266:  m.cs = 266; goto _test_eof
	_test_eof267:  m.cs = 267; goto _test_eof
	_test_eof268:  m.cs = 268; goto _test_eof
	_test_eof269:  m.cs = 269; goto _test_eof
	_test_eof270:  m.cs = 270; goto _test_eof
	_test_eof45:  m.cs = 45; goto _test_eof
	_test_eof46:  m.cs = 46; goto _test_eof
	_test_eof47:  m.cs = 47; goto _test_eof
	_test_eof271:  m.cs = 271; goto _test_eof
	_test_eof48:  m.cs = 48; goto _test_eof
	_test_eof49:  m.cs = 49; goto _test_eof
	_test_eof50:  m.cs = 50; goto _test_eof
	_test_eof51:  m.cs = 51; goto _test_eof
	_test_eof272:  m.cs = 272; goto _test_eof
	_test_eof273:  m.cs = 273; goto _test_eof
	_test_eof52:  m.cs = 52; goto _test_eof
	_test_eof274:  m.cs = 274; goto _test_eof
	_test_eof53:  m.cs = 53; goto _test_eof
	_test_eof275:  m.cs = 275; goto _test_eof
	_test_eof276:  m.cs = 276; goto _test_eof
	_test_eof277:  m.cs = 277; goto _test_eof
	_test_eof278:  m.cs = 278; goto _test_eof
	_test_eof279:  m.cs = 279; goto _test_eof
	_test_eof280:  m.cs = 280; goto _test_eof
	_test_eof281:  m.cs = 281; goto _test_eof
	_test_eof54:  m.cs = 54; goto _test_eof
	_test_eof55:  m.cs = 55; goto _test_eof
	_test_eof56:  m.cs = 56; goto _test_eof
	_test_eof282:  m.cs = 282; goto _test_eof
	_test_eof57:  m.cs = 57; goto _test_eof
	_test_eof58:  m.cs = 58; goto _test_eof
	_test_eof59:  m.cs = 59; goto _test_eof
	_test_eof283:  m.cs = 283; goto _test_eof
	_test_eof60:  m.cs = 60; goto _test_eof
	_test_eof61:  m.cs = 61; goto _test_eof
	_test_eof284:  m.cs = 284; goto _test_eof
	_test_eof285:  m.cs = 285; goto _test_eof
	_test_eof62:  m.cs = 62; goto _test_eof
	_test_eof63:  m.cs = 63; goto _test_eof
	_test_eof286:  m.cs = 286; goto _test_eof
	_test_eof287:  m.cs = 287; goto _test_eof
	_test_eof64:  m.cs = 64; goto _test_eof
	_test_eof65:  m.cs = 65; goto _test_eof
	_test_eof288:  m.cs = 288; goto _test_eof
	_test_eof289:  m.cs = 289; goto _test_eof
	_test_eof290:  m.cs = 290; goto _test_eof
	_test_eof291:  m.cs = 291; goto _test_eof
	_test_eof292:  m.cs = 292; goto _test_eof
	_test_eof293:  m.cs = 293; goto _test_eof
	_test_eof294:  m.cs = 294; goto _test_eof
	_test_eof66:  m.cs = 66; goto _test_eof
	_test_eof67:  m.cs = 67; goto _test_eof
	_test_eof68:  m.cs = 68; goto _test_eof
	_test_eof295:  m.cs = 295; goto _test_eof
	_test_eof69:  m.cs = 69; goto _test_eof
	_test_eof70:  m.cs = 70; goto _test_eof
	_test_eof71:  m.cs = 71; goto _test_eof
	_test_eof296:  m.cs = 296; goto _test_eof
	_test_eof72:  m.cs = 72; goto _test_eof
	_test_eof73:  m.cs = 73; goto _test_eof
	_test_eof297:  m.cs = 297; goto _test_eof
	_test_eof298:  m.cs = 298; goto _test_eof
	_test_eof74:  m.cs = 74; goto _test_eof
	_test_eof75:  m.cs = 75; goto _test_eof
	_test_eof76:  m.cs = 76; goto _test_eof
	_test_eof77:  m.cs = 77; goto _test_eof
	_test_eof78:  m.cs = 78; goto _test_eof
	_test_eof79:  m.cs = 79; goto _test_eof
	_test_eof299:  m.cs = 299; goto _test_eof
	_test_eof300:  m.cs = 300; goto _test_eof
	_test_eof301:  m.cs = 301; goto _test_eof
	_test_eof302:  m.cs = 302; goto _test_eof
	_test_eof80:  m.cs = 80; goto _test_eof
	_test_eof303:  m.cs = 303; goto _test_eof
	_test_eof304:  m.cs = 304; goto _test_eof
	_test_eof305:  m.cs = 305; goto _test_eof
	_test_eof306:  m.cs = 306; goto _test_eof
	_test_eof81:  m.cs = 81; goto _test_eof
	_test_eof307:  m.cs = 307; goto _test_eof
	_test_eof308:  m.cs = 308; goto _test_eof
	_test_eof309:  m.cs = 309; goto _test_eof
	_test_eof310:  m.cs = 310; goto _test_eof
	_test_eof311:  m.cs = 311; goto _test_eof
	_test_eof312:  m.cs = 312; goto _test_eof
	_test_eof313:  m.cs = 313; goto _test_eof
	_test_eof314:  m.cs = 314; goto _test_eof
	_test_eof315:  m.cs = 315; goto _test_eof
	_test_eof316:  m.cs = 316; goto _test_eof
	_test_eof317:  m.cs = 317; goto _test_eof
	_test_eof318:  m.cs = 318; goto _test_eof
	_test_eof319:  m.cs = 319; goto _test_eof
	_test_eof320:  m.cs = 320; goto _test_eof
	_test_eof321:  m.cs = 321; goto _test_eof
	_test_eof322:  m.cs = 322; goto _test_eof
	_test_eof323:  m.cs = 323; goto _test_eof
	_test_eof324:  m.cs = 324; goto _test_eof
	_test_eof82:  m.cs = 82; goto _test_eof
	_test_eof83:  m.cs = 83; goto _test_eof
	_test_eof84:  m.cs = 84; goto _test_eof
	_test_eof85:  m.cs = 85; goto _test_eof
	_test_eof86:  m.cs = 86; goto _test_eof
	_test_eof87:  m.cs = 87; goto _test_eof
	_test_eof88:  m.cs = 88; goto _test_eof
	_test_eof89:  m.cs = 89; goto _test_eof
	_test_eof90:  m.cs = 90; goto _test_eof
	_test_eof91:  m.cs = 91; goto _test_eof
	_test_eof92:  m.cs = 92; goto _test_eof
	_test_eof93:  m.cs = 93; goto _test_eof
	_test_eof94:  m.cs = 94; goto _test_eof
	_test_eof325:  m.cs = 325; goto _test_eof
	_test_eof326:  m.cs = 326; goto _test_eof
	_test_eof95:  m.cs = 95; goto _test_eof
	_test_eof327:  m.cs = 327; goto _test_eof
	_test_eof328:  m.cs = 328; goto _test_eof
	_test_eof329:  m.cs = 329; goto _test_eof
	_test_eof330:  m.cs = 330; goto _test_eof
	_test_eof331:  m.cs = 331; goto _test_eof
	_test_eof332:  m.cs = 332; goto _test_eof
	_test_eof333:  m.cs = 333; goto _test_eof
	_test_eof334:  m.cs = 334; goto _test_eof
	_test_eof335:  m.cs = 335; goto _test_eof
	_test_eof336:  m.cs = 336; goto _test_eof
	_test_eof337:  m.cs = 337; goto _test_eof
	_test_eof338:  m.cs = 338; goto _test_eof
	_test_eof339:  m.cs = 339; goto _test_eof
	_test_eof340:  m.cs = 340; goto _test_eof
	_test_eof341:  m.cs = 341; goto _test_eof
	_test_eof342:  m.cs = 342; goto _test_eof
	_test_eof343:  m.cs = 343; goto _test_eof
	_test_eof344:  m.cs = 344; goto _test_eof
	_test_eof345:  m.cs = 345; goto _test_eof
	_test_eof346:  m.cs = 346; goto _test_eof
	_test_eof96:  m.cs = 96; goto _test_eof
	_test_eof97:  m.cs = 97; goto _test_eof
	_test_eof347:  m.cs = 347; goto _test_eof
	_test_eof348:  m.cs = 348; goto _test_eof
	_test_eof98:  m.cs = 98; goto _test_eof
	_test_eof349:  m.cs = 349; goto _test_eof
	_test_eof350:  m.cs = 350; goto _test_eof
	_test_eof351:  m.cs = 351; goto _test_eof
	_test_eof352:  m.cs = 352; goto _test_eof
	_test_eof353:  m.cs = 353; goto _test_eof
	_test_eof354:  m.cs = 354; goto _test_eof
	_test_eof355:  m.cs = 355; goto _test_eof
	_test_eof356:  m.cs = 356; goto _test_eof
	_test_eof357:  m.cs = 357; goto _test_eof
	_test_eof358:  m.cs = 358; goto _test_eof
	_test_eof359:  m.cs = 359; goto _test_eof
	_test_eof360:  m.cs = 360; goto _test_eof
	_test_eof361:  m.cs = 361; goto _test_eof
	_test_eof362:  m.cs = 362; goto _test_eof
	_test_eof363:  m.cs = 363; goto _test_eof
	_test_eof364:  m.cs = 364; goto _test_eof
	_test_eof365:  m.cs = 365; goto _test_eof
	_test_eof366:  m.cs = 366; goto _test_eof
	_test_eof367:  m.cs = 367; goto _test_eof
	_test_eof368:  m.cs = 368; goto _test_eof
	_test_eof99:  m.cs = 99; goto _test_eof
	_test_eof369:  m.cs = 369; goto _test_eof
	_test_eof370:  m.cs = 370; goto _test_eof
	_test_eof100:  m.cs = 100; goto _test_eof
	_test_eof101:  m.cs = 101; goto _test_eof
	_test_eof102:  m.cs = 102; goto _test_eof
	_test_eof103:  m.cs = 103; goto _test_eof
	_test_eof371:  m.cs = 371; goto _test_eof
	_test_eof372:  m.cs = 372; goto _test_eof
	_test_eof104:  m.cs = 104; goto _test_eof
	_test_eof105:  m.cs = 105; goto _test_eof
	_test_eof373:  m.cs = 373; goto _test_eof
	_test_eof374:  m.cs = 374; goto _test_eof
	_test_eof375:  m.cs = 375; goto _test_eof
	_test_eof376:  m.cs = 376; goto _test_eof
	_test_eof377:  m.cs = 377; goto _test_eof
	_test_eof378:  m.cs = 378; goto _test_eof
	_test_eof379:  m.cs = 379; goto _test_eof
	_test_eof106:  m.cs = 106; goto _test_eof
	_test_eof107:  m.cs = 107; goto _test_eof
	_test_eof108:  m.cs = 108; goto _test_eof
	_test_eof380:  m.cs = 380; goto _test_eof
	_test_eof109:  m.cs = 109; goto _test_eof
	_test_eof110:  m.cs = 110; goto _test_eof
	_test_eof111:  m.cs = 111; goto _test_eof
	_test_eof381:  m.cs = 381; goto _test_eof
	_test_eof112:  m.cs = 112; goto _test_eof
	_test_eof113:  m.cs = 113; goto _test_eof
	_test_eof382:  m.cs = 382; goto _test_eof
	_test_eof383:  m.cs = 383; goto _test_eof
	_test_eof114:  m.cs = 114; goto _test_eof
	_test_eof115:  m.cs = 115; goto _test_eof
	_test_eof116:  m.cs = 116; goto _test_eof
	_test_eof117:  m.cs = 117; goto _test_eof
	_test_eof118:  m.cs = 118; goto _test_eof
	_test_eof119:  m.cs = 119; goto _test_eof
	_test_eof120:  m.cs = 120; goto _test_eof
	_test_eof121:  m.cs = 121; goto _test_eof
	_test_eof122:  m.cs = 122; goto _test_eof
	_test_eof123:  m.cs = 123; goto _test_eof
	_test_eof124:  m.cs = 124; goto _test_eof
	_test_eof125:  m.cs = 125; goto _test_eof
	_test_eof384:  m.cs = 384; goto _test_eof
	_test_eof385:  m.cs = 385; goto _test_eof
	_test_eof386:  m.cs = 386; goto _test_eof
	_test_eof126:  m.cs = 126; goto _test_eof
	_test_eof387:  m.cs = 387; goto _test_eof
	_test_eof388:  m.cs = 388; goto _test_eof
	_test_eof389:  m.cs = 389; goto _test_eof
	_test_eof390:  m.cs = 390; goto _test_eof
	_test_eof391:  m.cs = 391; goto _test_eof
	_test_eof392:  m.cs = 392; goto _test_eof
	_test_eof393:  m.cs = 393; goto _test_eof
	_test_eof394:  m.cs = 394; goto _test_eof
	_test_eof395:  m.cs = 395; goto _test_eof
	_test_eof396:  m.cs = 396; goto _test_eof
	_test_eof397:  m.cs = 397; goto _test_eof
	_test_eof398:  m.cs = 398; goto _test_eof
	_test_eof399:  m.cs = 399; goto _test_eof
	_test_eof400:  m.cs = 400; goto _test_eof
	_test_eof401:  m.cs = 401; goto _test_eof
	_test_eof402:  m.cs = 402; goto _test_eof
	_test_eof403:  m.cs = 403; goto _test_eof
	_test_eof404:  m.cs = 404; goto _test_eof
	_test_eof405:  m.cs = 405; goto _test_eof
	_test_eof406:  m.cs = 406; goto _test_eof
	_test_eof407:  m.cs = 407; goto _test_eof
	_test_eof408:  m.cs = 408; goto _test_eof
	_test_eof127:  m.cs = 127; goto _test_eof
	_test_eof409:  m.cs = 409; goto _test_eof
	_test_eof410:  m.cs = 410; goto _test_eof
	_test_eof411:  m.cs = 411; goto _test_eof
	_test_eof412:  m.cs = 412; goto _test_eof
	_test_eof128:  m.cs = 128; goto _test_eof
	_test_eof413:  m.cs = 413; goto _test_eof
	_test_eof414:  m.cs = 414; goto _test_eof
	_test_eof415:  m.cs = 415; goto _test_eof
	_test_eof416:  m.cs = 416; goto _test_eof
	_test_eof417:  m.cs = 417; goto _test_eof
	_test_eof418:  m.cs = 418; goto _test_eof
	_test_eof419:  m.cs = 419; goto _test_eof
	_test_eof420:  m.cs = 420; goto _test_eof
	_test_eof421:  m.cs = 421; goto _test_eof
	_test_eof422:  m.cs = 422; goto _test_eof
	_test_eof423:  m.cs = 423; goto _test_eof
	_test_eof424:  m.cs = 424; goto _test_eof
	_test_eof425:  m.cs = 425; goto _test_eof
	_test_eof426:  m.cs = 426; goto _test_eof
	_test_eof427:  m.cs = 427; goto _test_eof
	_test_eof428:  m.cs = 428; goto _test_eof
	_test_eof429:  m.cs = 429; goto _test_eof
	_test_eof430:  m.cs = 430; goto _test_eof
	_test_eof431:  m.cs = 431; goto _test_eof
	_test_eof432:  m.cs = 432; goto _test_eof
	_test_eof129:  m.cs = 129; goto _test_eof
	_test_eof130:  m.cs = 130; goto _test_eof
	_test_eof131:  m.cs = 131; goto _test_eof
	_test_eof433:  m.cs = 433; goto _test_eof
	_test_eof434:  m.cs = 434; goto _test_eof
	_test_eof132:  m.cs = 132; goto _test_eof
	_test_eof435:  m.cs = 435; goto _test_eof
	_test_eof436:  m.cs = 436; goto _test_eof
	_test_eof437:  m.cs = 437; goto _test_eof
	_test_eof438:  m.cs = 438; goto _test_eof
	_test_eof439:  m.cs = 439; goto _test_eof
	_test_eof440:  m.cs = 440; goto _test_eof
	_test_eof441:  m.cs = 441; goto _test_eof
	_test_eof442:  m.cs = 442; goto _test_eof
	_test_eof443:  m.cs = 443; goto _test_eof
	_test_eof444:  m.cs = 444; goto _test_eof
	_test_eof445:  m.cs = 445; goto _test_eof
	_test_eof446:  m.cs = 446; goto _test_eof
	_test_eof447:  m.cs = 447; goto _test_eof
	_test_eof448:  m.cs = 448; goto _test_eof
	_test_eof449:  m.cs = 449; goto _test_eof
	_test_eof450:  m.cs = 450; goto _test_eof
	_test_eof451:  m.cs = 451; goto _test_eof
	_test_eof452:  m.cs = 452; goto _test_eof
	_test_eof453:  m.cs = 453; goto _test_eof
	_test_eof454:  m.cs = 454; goto _test_eof
	_test_eof133:  m.cs = 133; goto _test_eof
	_test_eof455:  m.cs = 455; goto _test_eof
	_test_eof456:  m.cs = 456; goto _test_eof
	_test_eof457:  m.cs = 457; goto _test_eof
	_test_eof134:  m.cs = 134; goto _test_eof
	_test_eof458:  m.cs = 458; goto _test_eof
	_test_eof459:  m.cs = 459; goto _test_eof
	_test_eof460:  m.cs = 460; goto _test_eof
	_test_eof461:  m.cs = 461; goto _test_eof
	_test_eof462:  m.cs = 462; goto _test_eof
	_test_eof463:  m.cs = 463; goto _test_eof
	_test_eof464:  m.cs = 464; goto _test_eof
	_test_eof465:  m.cs = 465; goto _test_eof
	_test_eof466:  m.cs = 466; goto _test_eof
	_test_eof467:  m.cs = 467; goto _test_eof
	_test_eof468:  m.cs = 468; goto _test_eof
	_test_eof469:  m.cs = 469; goto _test_eof
	_test_eof470:  m.cs = 470; goto _test_eof
	_test_eof471:  m.cs = 471; goto _test_eof
	_test_eof472:  m.cs = 472; goto _test_eof
	_test_eof473:  m.cs = 473; goto _test_eof
	_test_eof474:  m.cs = 474; goto _test_eof
	_test_eof475:  m.cs = 475; goto _test_eof
	_test_eof476:  m.cs = 476; goto _test_eof
	_test_eof477:  m.cs = 477; goto _test_eof
	_test_eof478:  m.cs = 478; goto _test_eof
	_test_eof479:  m.cs = 479; goto _test_eof
	_test_eof135:  m.cs = 135; goto _test_eof
	_test_eof480:  m.cs = 480; goto _test_eof
	_test_eof481:  m.cs = 481; goto _test_eof
	_test_eof482:  m.cs = 482; goto _test_eof
	_test_eof483:  m.cs = 483; goto _test_eof
	_test_eof484:  m.cs = 484; goto _test_eof
	_test_eof485:  m.cs = 485; goto _test_eof
	_test_eof486:  m.cs = 486; goto _test_eof
	_test_eof487:  m.cs = 487; goto _test_eof
	_test_eof488:  m.cs = 488; goto _test_eof
	_test_eof489:  m.cs = 489; goto _test_eof
	_test_eof490:  m.cs = 490; goto _test_eof
	_test_eof491:  m.cs = 491; goto _test_eof
	_test_eof492:  m.cs = 492; goto _test_eof
	_test_eof493:  m.cs = 493; goto _test_eof
	_test_eof494:  m.cs = 494; goto _test_eof
	_test_eof495:  m.cs = 495; goto _test_eof
	_test_eof496:  m.cs = 496; goto _test_eof
	_test_eof497:  m.cs = 497; goto _test_eof
	_test_eof498:  m.cs = 498; goto _test_eof
	_test_eof499:  m.cs = 499; goto _test_eof
	_test_eof500:  m.cs = 500; goto _test_eof
	_test_eof501:  m.cs = 501; goto _test_eof
	_test_eof136:  m.cs = 136; goto _test_eof
	_test_eof137:  m.cs = 137; goto _test_eof
	_test_eof138:  m.cs = 138; goto _test_eof
	_test_eof139:  m.cs = 139; goto _test_eof
	_test_eof502:  m.cs = 502; goto _test_eof
	_test_eof503:  m.cs = 503; goto _test_eof
	_test_eof140:  m.cs = 140; goto _test_eof
	_test_eof504:  m.cs = 504; goto _test_eof
	_test_eof141:  m.cs = 141; goto _test_eof
	_test_eof505:  m.cs = 505; goto _test_eof
	_test_eof506:  m.cs = 506; goto _test_eof
	_test_eof507:  m.cs = 507; goto _test_eof
	_test_eof508:  m.cs = 508; goto _test_eof
	_test_eof509:  m.cs = 509; goto _test_eof
	_test_eof510:  m.cs = 510; goto _test_eof
	_test_eof511:  m.cs = 511; goto _test_eof
	_test_eof142:  m.cs = 142; goto _test_eof
	_test_eof143:  m.cs = 143; goto _test_eof
	_test_eof144:  m.cs = 144; goto _test_eof
	_test_eof512:  m.cs = 512; goto _test_eof
	_test_eof145:  m.cs = 145; goto _test_eof
	_test_eof146:  m.cs = 146; goto _test_eof
	_test_eof147:  m.cs = 147; goto _test_eof
	_test_eof513:  m.cs = 513; goto _test_eof
	_test_eof148:  m.cs = 148; goto _test_eof
	_test_eof149:  m.cs = 149; goto _test_eof
	_test_eof514:  m.cs = 514; goto _test_eof
	_test_eof515:  m.cs = 515; goto _test_eof
	_test_eof150:  m.cs = 150; goto _test_eof
	_test_eof151:  m.cs = 151; goto _test_eof
	_test_eof516:  m.cs = 516; goto _test_eof
	_test_eof517:  m.cs = 517; goto _test_eof
	_test_eof518:  m.cs = 518; goto _test_eof
	_test_eof152:  m.cs = 152; goto _test_eof
	_test_eof519:  m.cs = 519; goto _test_eof
	_test_eof520:  m.cs = 520; goto _test_eof
	_test_eof521:  m.cs = 521; goto _test_eof
	_test_eof522:  m.cs = 522; goto _test_eof
	_test_eof523:  m.cs = 523; goto _test_eof
	_test_eof524:  m.cs = 524; goto _test_eof
	_test_eof525:  m.cs = 525; goto _test_eof
	_test_eof526:  m.cs = 526; goto _test_eof
	_test_eof527:  m.cs = 527; goto _test_eof
	_test_eof528:  m.cs = 528; goto _test_eof
	_test_eof529:  m.cs = 529; goto _test_eof
	_test_eof530:  m.cs = 530; goto _test_eof
	_test_eof531:  m.cs = 531; goto _test_eof
	_test_eof532:  m.cs = 532; goto _test_eof
	_test_eof533:  m.cs = 533; goto _test_eof
	_test_eof534:  m.cs = 534; goto _test_eof
	_test_eof535:  m.cs = 535; goto _test_eof
	_test_eof536:  m.cs = 536; goto _test_eof
	_test_eof537:  m.cs = 537; goto _test_eof
	_test_eof538:  m.cs = 538; goto _test_eof
	_test_eof539:  m.cs = 539; goto _test_eof
	_test_eof153:  m.cs = 153; goto _test_eof
	_test_eof154:  m.cs = 154; goto _test_eof
	_test_eof540:  m.cs = 540; goto _test_eof
	_test_eof541:  m.cs = 541; goto _test_eof
	_test_eof542:  m.cs = 542; goto _test_eof
	_test_eof543:  m.cs = 543; goto _test_eof
	_test_eof544:  m.cs = 544; goto _test_eof
	_test_eof545:  m.cs = 545; goto _test_eof
	_test_eof546:  m.cs = 546; goto _test_eof
	_test_eof155:  m.cs = 155; goto _test_eof
	_test_eof156:  m.cs = 156; goto _test_eof
	_test_eof157:  m.cs = 157; goto _test_eof
	_test_eof547:  m.cs = 547; goto _test_eof
	_test_eof158:  m.cs = 158; goto _test_eof
	_test_eof159:  m.cs = 159; goto _test_eof
	_test_eof160:  m.cs = 160; goto _test_eof
	_test_eof548:  m.cs = 548; goto _test_eof
	_test_eof161:  m.cs = 161; goto _test_eof
	_test_eof162:  m.cs = 162; goto _test_eof
	_test_eof549:  m.cs = 549; goto _test_eof
	_test_eof550:  m.cs = 550; goto _test_eof
	_test_eof163:  m.cs = 163; goto _test_eof
	_test_eof164:  m.cs = 164; goto _test_eof
	_test_eof165:  m.cs = 165; goto _test_eof
	_test_eof551:  m.cs = 551; goto _test_eof
	_test_eof552:  m.cs = 552; goto _test_eof
	_test_eof166:  m.cs = 166; goto _test_eof
	_test_eof553:  m.cs = 553; goto _test_eof
	_test_eof554:  m.cs = 554; goto _test_eof
	_test_eof167:  m.cs = 167; goto _test_eof
	_test_eof555:  m.cs = 555; goto _test_eof
	_test_eof556:  m.cs = 556; goto _test_eof
	_test_eof557:  m.cs = 557; goto _test_eof
	_test_eof558:  m.cs = 558; goto _test_eof
	_test_eof559:  m.cs = 559; goto _test_eof
	_test_eof560:  m.cs = 560; goto _test_eof
	_test_eof561:  m.cs = 561; goto _test_eof
	_test_eof168:  m.cs = 168; goto _test_eof
	_test_eof169:  m.cs = 169; goto _test_eof
	_test_eof170:  m.cs = 170; goto _test_eof
	_test_eof562:  m.cs = 562; goto _test_eof
	_test_eof171:  m.cs = 171; goto _test_eof
	_test_eof172:  m.cs = 172; goto _test_eof
	_test_eof173:  m.cs = 173; goto _test_eof
	_test_eof563:  m.cs = 563; goto _test_eof
	_test_eof174:  m.cs = 174; goto _test_eof
	_test_eof175:  m.cs = 175; goto _test_eof
	_test_eof564:  m.cs = 564; goto _test_eof
	_test_eof565:  m.cs = 565; goto _test_eof
	_test_eof176:  m.cs = 176; goto _test_eof
	_test_eof566:  m.cs = 566; goto _test_eof
	_test_eof567:  m.cs = 567; goto _test_eof
	_test_eof177:  m.cs = 177; goto _test_eof
	_test_eof178:  m.cs = 178; goto _test_eof
	_test_eof568:  m.cs = 568; goto _test_eof
	_test_eof569:  m.cs = 569; goto _test_eof
	_test_eof570:  m.cs = 570; goto _test_eof
	_test_eof571:  m.cs = 571; goto _test_eof
	_test_eof572:  m.cs = 572; goto _test_eof
	_test_eof573:  m.cs = 573; goto _test_eof
	_test_eof179:  m.cs = 179; goto _test_eof
	_test_eof180:  m.cs = 180; goto _test_eof
	_test_eof181:  m.cs = 181; goto _test_eof
	_test_eof574:  m.cs = 574; goto _test_eof
	_test_eof182:  m.cs = 182; goto _test_eof
	_test_eof183:  m.cs = 183; goto _test_eof
	_test_eof184:  m.cs = 184; goto _test_eof
	_test_eof575:  m.cs = 575; goto _test_eof
	_test_eof185:  m.cs = 185; goto _test_eof
	_test_eof186:  m.cs = 186; goto _test_eof
	_test_eof576:  m.cs = 576; goto _test_eof
	_test_eof577:  m.cs = 577; goto _test_eof
	_test_eof187:  m.cs = 187; goto _test_eof
	_test_eof578:  m.cs = 578; goto _test_eof
	_test_eof188:  m.cs = 188; goto _test_eof
	_test_eof579:  m.cs = 579; goto _test_eof
	_test_eof580:  m.cs = 580; goto _test_eof
	_test_eof189:  m.cs = 189; goto _test_eof
	_test_eof190:  m.cs = 190; goto _test_eof

	_test_eof: {}
	if ( m.p) == ( m.eof) {
		switch  m.cs {
		case 191, 192, 193, 195, 227, 228, 230, 249, 250, 252, 271, 274, 299, 300, 301, 302, 304, 305, 306, 325, 326, 328, 347, 348, 350, 369, 370, 384, 385, 386, 388, 407, 408, 410, 411, 412, 431, 432, 433, 434, 436, 455, 456, 457, 459, 478, 479, 481, 482, 483, 504, 517, 518, 520, 553, 579, 580:
//line plugins/parsers/influx/machine.go.rl:22

	yield = true
	 m.cs = 188;
	{( m.p)++;  m.cs = 0; goto _out }

		case 1, 129:
//line plugins/parsers/influx/machine.go.rl:56

	m.err = ErrParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++;  m.cs = 0; goto _out }

		case 2, 3, 4, 5, 6, 7, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 37, 38, 39, 40, 41, 43, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 81, 87, 88, 89, 90, 91, 125, 128, 164, 165, 166, 167, 168, 169, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 180, 181, 182, 183, 184, 185, 186:
//line plugins/parsers/influx/machine.go.rl:35

	m.err = ErrFieldParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++;  m.cs = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:56

	m.err = ErrParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++;  m.cs = 0; goto _out }

		case 27, 28, 29, 35, 36:
//line plugins/parsers/influx/machine.go.rl:42

	m.err = ErrTagParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++;  m.cs = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:56

	m.err = ErrParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++;  m.cs = 0; goto _out }

		case 8:
//line plugins/parsers/influx/machine.go.rl:49

	m.err = ErrTimestampParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++;  m.cs = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:56

	m.err = ErrParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++;  m.cs = 0; goto _out }

		case 217, 276, 289, 374, 506, 541, 556, 568:
//line plugins/parsers/influx/machine.go.rl:88

	m.handler.AddInt(key, m.text())

//line plugins/parsers/influx/machine.go.rl:22

	yield = true
	 m.cs = 188;
	{( m.p)++;  m.cs = 0; goto _out }

		case 220, 279, 292, 377, 509, 544, 559, 571:
//line plugins/parsers/influx/machine.go.rl:92

	m.handler.AddUint(key, m.text())

//line plugins/parsers/influx/machine.go.rl:22

	yield = true
	 m.cs = 188;
	{( m.p)++;  m.cs = 0; goto _out }

		case 214, 215, 216, 218, 219, 221, 272, 273, 275, 277, 278, 280, 286, 287, 288, 290, 291, 293, 371, 372, 373, 375, 376, 378, 502, 503, 505, 507, 508, 510, 516, 539, 540, 542, 543, 545, 551, 552, 554, 555, 557, 558, 560, 566, 567, 569, 570, 572:
//line plugins/parsers/influx/machine.go.rl:96

	m.handler.AddFloat(key, m.text())

//line plugins/parsers/influx/machine.go.rl:22

	yield = true
	 m.cs = 188;
	{( m.p)++;  m.cs = 0; goto _out }

		case 222, 223, 224, 225, 226, 281, 282, 283, 284, 285, 294, 295, 296, 297, 298, 379, 380, 381, 382, 383, 511, 512, 513, 514, 515, 546, 547, 548, 549, 550, 561, 562, 563, 564, 565, 573, 574, 575, 576, 577:
//line plugins/parsers/influx/machine.go.rl:100

	m.handler.AddBool(key, m.text())

//line plugins/parsers/influx/machine.go.rl:22

	yield = true
	 m.cs = 188;
	{( m.p)++;  m.cs = 0; goto _out }

		case 194, 196, 197, 198, 199, 200, 201, 202, 203, 204, 205, 206, 207, 208, 209, 210, 211, 212, 213, 229, 231, 232, 233, 234, 235, 236, 237, 238, 239, 240, 241, 242, 243, 244, 245, 246, 247, 248, 251, 253, 254, 255, 256, 257, 258, 259, 260, 261, 262, 263, 264, 265, 266, 267, 268, 269, 270, 303, 307, 308, 309, 310, 311, 312, 313, 314, 315, 316, 317, 318, 319, 320, 321, 322, 323, 324, 327, 329, 330, 331, 332, 333, 334, 335, 336, 337, 338, 339, 340, 341, 342, 343, 344, 345, 346, 349, 351, 352, 353, 354, 355, 356, 357, 358, 359, 360, 361, 362, 363, 364, 365, 366, 367, 368, 387, 389, 390, 391, 392, 393, 394, 395, 396, 397, 398, 399, 400, 401, 402, 403, 404, 405, 406, 409, 413, 414, 415, 416, 417, 418, 419, 420, 421, 422, 423, 424, 425, 426, 427, 428, 429, 430, 435, 437, 438, 439, 440, 441, 442, 443, 444, 445, 446, 447, 448, 449, 450, 451, 452, 453, 454, 458, 460, 461, 462, 463, 464, 465, 466, 467, 468, 469, 470, 471, 472, 473, 474, 475, 476, 477, 480, 484, 485, 486, 487, 488, 489, 490, 491, 492, 493, 494, 495, 496, 497, 498, 499, 500, 501, 519, 521, 522, 523, 524, 525, 526, 527, 528, 529, 530, 531, 532, 533, 534, 535, 536, 537, 538:
//line plugins/parsers/influx/machine.go.rl:108

	m.handler.SetTimestamp(m.text())

//line plugins/parsers/influx/machine.go.rl:22

	yield = true
	 m.cs = 188;
	{( m.p)++;  m.cs = 0; goto _out }

		case 42, 44, 80, 126, 127, 134:
//line plugins/parsers/influx/machine.go.rl:35

	m.err = ErrFieldParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++;  m.cs = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:49

	m.err = ErrTimestampParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++;  m.cs = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:56

	m.err = ErrParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++;  m.cs = 0; goto _out }

		case 30, 31, 32, 33, 34, 82, 83, 84, 85, 86, 92, 93, 94, 96, 97, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 130, 131, 133, 136, 137, 138, 139, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 150, 151, 153, 154, 155, 156, 157, 158, 159, 160, 161, 162, 163:
//line plugins/parsers/influx/machine.go.rl:42

	m.err = ErrTagParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++;  m.cs = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:35

	m.err = ErrFieldParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++;  m.cs = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:56

	m.err = ErrParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++;  m.cs = 0; goto _out }

		case 98:
//line plugins/parsers/influx/machine.go.rl:42

	m.err = ErrTagParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++;  m.cs = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:49

	m.err = ErrTimestampParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++;  m.cs = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:56

	m.err = ErrParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++;  m.cs = 0; goto _out }

		case 95, 132, 135, 152:
//line plugins/parsers/influx/machine.go.rl:42

	m.err = ErrTagParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++;  m.cs = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:35

	m.err = ErrFieldParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++;  m.cs = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:49

	m.err = ErrTimestampParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++;  m.cs = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:56

	m.err = ErrParse
	( m.p)--

	 m.cs = 187;
	{( m.p)++;  m.cs = 0; goto _out }

//line plugins/parsers/influx/machine.go:22402
		}
	}

	_out: {}
	}

//line plugins/parsers/influx/machine.go.rl:288

	// Even if there was an error, return true. On the next call to this
	// function we will attempt to scan to the next line of input and recover.
	if m.err != nil {
		return true
	}

	// Don't check the error state in the case that we just yielded, because
	// the yield indicates we just completed parsing a line.
	if !yield && m.cs == LineProtocol_error {
		m.err = ErrParse
		return true
	}

	return true
}

// Err returns the error that occurred on the last call to ParseLine.  If the
// result is nil, then the line was parsed successfully.
func (m *machine) Err() error {
	return m.err
}

// Position returns the current position into the input.
func (m *machine) Position() int {
	return m.p
}

func (m *machine) text() []byte {
	return m.data[m.pb:m.p]
}
