
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
	EOF = errors.New("EOF")
)


//line plugins/parsers/influx/machine.go.rl:304



//line plugins/parsers/influx/machine.go:24
const LineProtocol_start int = 259
const LineProtocol_first_final int = 259
const LineProtocol_error int = 0

const LineProtocol_en_main int = 259
const LineProtocol_en_discard_line int = 247
const LineProtocol_en_align int = 715
const LineProtocol_en_series int = 250


//line plugins/parsers/influx/machine.go.rl:307

type Handler interface {
	SetMeasurement(name []byte) error
	AddTag(key []byte, value []byte) error
	AddInt(key []byte, value []byte) error
	AddUint(key []byte, value []byte) error
	AddFloat(key []byte, value []byte) error
	AddString(key []byte, value []byte) error
	AddBool(key []byte, value []byte) error
	SetTimestamp(tm []byte) error
}

type machine struct {
	data       []byte
	cs         int
	p, pe, eof int
	pb         int
	lineno     int
	sol        int
	handler    Handler
	initState  int
}

func NewMachine(handler Handler) *machine {
	m := &machine{
		handler: handler,
		initState: LineProtocol_en_align,
	}

	
//line plugins/parsers/influx/machine.go.rl:337
	
//line plugins/parsers/influx/machine.go.rl:338
	
//line plugins/parsers/influx/machine.go.rl:339
	
//line plugins/parsers/influx/machine.go.rl:340
	
//line plugins/parsers/influx/machine.go.rl:341
	
//line plugins/parsers/influx/machine.go.rl:342
	
//line plugins/parsers/influx/machine.go:78
	{
	( m.cs) = LineProtocol_start
	}

//line plugins/parsers/influx/machine.go.rl:343

	return m
}

func NewSeriesMachine(handler Handler) *machine {
	m := &machine{
		handler: handler,
		initState: LineProtocol_en_series,
	}

	
//line plugins/parsers/influx/machine.go.rl:354
	
//line plugins/parsers/influx/machine.go.rl:355
	
//line plugins/parsers/influx/machine.go.rl:356
	
//line plugins/parsers/influx/machine.go.rl:357
	
//line plugins/parsers/influx/machine.go.rl:358
	
//line plugins/parsers/influx/machine.go:105
	{
	( m.cs) = LineProtocol_start
	}

//line plugins/parsers/influx/machine.go.rl:359

	return m
}

func (m *machine) SetData(data []byte) {
	m.data = data
	m.p = 0
	m.pb = 0
	m.lineno = 1
	m.sol = 0
	m.pe = len(data)
	m.eof = len(data)

	
//line plugins/parsers/influx/machine.go:125
	{
	( m.cs) = LineProtocol_start
	}

//line plugins/parsers/influx/machine.go.rl:373
	m.cs = m.initState
}

// Next parses the next metric line and returns nil if it was successfully
// processed.  If the line contains a syntax error an error is returned,
// otherwise if the end of file is reached before finding a metric line then
// EOF is returned.
func (m *machine) Next() error {
	if m.p == m.pe && m.pe == m.eof {
		return EOF
	}

	var err error
	var key []byte
	foundMetric := false

	
//line plugins/parsers/influx/machine.go:148
	{
	if ( m.p) == ( m.pe) {
		goto _test_eof
	}
	goto _resume

_again:
	switch ( m.cs) {
	case 259:
		goto st259
	case 1:
		goto st1
	case 2:
		goto st2
	case 3:
		goto st3
	case 0:
		goto st0
	case 4:
		goto st4
	case 5:
		goto st5
	case 6:
		goto st6
	case 7:
		goto st7
	case 8:
		goto st8
	case 260:
		goto st260
	case 261:
		goto st261
	case 262:
		goto st262
	case 9:
		goto st9
	case 10:
		goto st10
	case 11:
		goto st11
	case 12:
		goto st12
	case 13:
		goto st13
	case 14:
		goto st14
	case 15:
		goto st15
	case 16:
		goto st16
	case 17:
		goto st17
	case 18:
		goto st18
	case 19:
		goto st19
	case 20:
		goto st20
	case 21:
		goto st21
	case 22:
		goto st22
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
	case 263:
		goto st263
	case 264:
		goto st264
	case 34:
		goto st34
	case 35:
		goto st35
	case 265:
		goto st265
	case 266:
		goto st266
	case 267:
		goto st267
	case 36:
		goto st36
	case 268:
		goto st268
	case 269:
		goto st269
	case 270:
		goto st270
	case 271:
		goto st271
	case 272:
		goto st272
	case 273:
		goto st273
	case 274:
		goto st274
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
	case 282:
		goto st282
	case 283:
		goto st283
	case 284:
		goto st284
	case 285:
		goto st285
	case 37:
		goto st37
	case 38:
		goto st38
	case 286:
		goto st286
	case 287:
		goto st287
	case 288:
		goto st288
	case 39:
		goto st39
	case 40:
		goto st40
	case 41:
		goto st41
	case 42:
		goto st42
	case 43:
		goto st43
	case 289:
		goto st289
	case 290:
		goto st290
	case 291:
		goto st291
	case 292:
		goto st292
	case 44:
		goto st44
	case 293:
		goto st293
	case 294:
		goto st294
	case 295:
		goto st295
	case 296:
		goto st296
	case 297:
		goto st297
	case 298:
		goto st298
	case 299:
		goto st299
	case 300:
		goto st300
	case 301:
		goto st301
	case 302:
		goto st302
	case 303:
		goto st303
	case 304:
		goto st304
	case 305:
		goto st305
	case 306:
		goto st306
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
	case 45:
		goto st45
	case 46:
		goto st46
	case 47:
		goto st47
	case 48:
		goto st48
	case 49:
		goto st49
	case 50:
		goto st50
	case 51:
		goto st51
	case 52:
		goto st52
	case 53:
		goto st53
	case 54:
		goto st54
	case 315:
		goto st315
	case 316:
		goto st316
	case 317:
		goto st317
	case 55:
		goto st55
	case 56:
		goto st56
	case 57:
		goto st57
	case 58:
		goto st58
	case 59:
		goto st59
	case 60:
		goto st60
	case 318:
		goto st318
	case 319:
		goto st319
	case 61:
		goto st61
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
	case 325:
		goto st325
	case 326:
		goto st326
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
	case 62:
		goto st62
	case 340:
		goto st340
	case 341:
		goto st341
	case 342:
		goto st342
	case 63:
		goto st63
	case 343:
		goto st343
	case 344:
		goto st344
	case 345:
		goto st345
	case 346:
		goto st346
	case 347:
		goto st347
	case 348:
		goto st348
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
	case 64:
		goto st64
	case 65:
		goto st65
	case 66:
		goto st66
	case 67:
		goto st67
	case 68:
		goto st68
	case 363:
		goto st363
	case 69:
		goto st69
	case 70:
		goto st70
	case 71:
		goto st71
	case 72:
		goto st72
	case 73:
		goto st73
	case 364:
		goto st364
	case 365:
		goto st365
	case 366:
		goto st366
	case 74:
		goto st74
	case 75:
		goto st75
	case 367:
		goto st367
	case 368:
		goto st368
	case 76:
		goto st76
	case 369:
		goto st369
	case 77:
		goto st77
	case 370:
		goto st370
	case 371:
		goto st371
	case 372:
		goto st372
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
	case 380:
		goto st380
	case 381:
		goto st381
	case 382:
		goto st382
	case 383:
		goto st383
	case 384:
		goto st384
	case 385:
		goto st385
	case 386:
		goto st386
	case 387:
		goto st387
	case 388:
		goto st388
	case 389:
		goto st389
	case 78:
		goto st78
	case 79:
		goto st79
	case 80:
		goto st80
	case 81:
		goto st81
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
	case 390:
		goto st390
	case 391:
		goto st391
	case 392:
		goto st392
	case 393:
		goto st393
	case 92:
		goto st92
	case 93:
		goto st93
	case 94:
		goto st94
	case 95:
		goto st95
	case 394:
		goto st394
	case 395:
		goto st395
	case 96:
		goto st96
	case 97:
		goto st97
	case 396:
		goto st396
	case 98:
		goto st98
	case 99:
		goto st99
	case 397:
		goto st397
	case 398:
		goto st398
	case 100:
		goto st100
	case 399:
		goto st399
	case 400:
		goto st400
	case 101:
		goto st101
	case 102:
		goto st102
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
	case 409:
		goto st409
	case 410:
		goto st410
	case 411:
		goto st411
	case 412:
		goto st412
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
	case 103:
		goto st103
	case 419:
		goto st419
	case 420:
		goto st420
	case 421:
		goto st421
	case 104:
		goto st104
	case 105:
		goto st105
	case 422:
		goto st422
	case 423:
		goto st423
	case 424:
		goto st424
	case 106:
		goto st106
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
	case 433:
		goto st433
	case 434:
		goto st434
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
	case 107:
		goto st107
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
	case 455:
		goto st455
	case 456:
		goto st456
	case 457:
		goto st457
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
	case 108:
		goto st108
	case 109:
		goto st109
	case 110:
		goto st110
	case 111:
		goto st111
	case 112:
		goto st112
	case 467:
		goto st467
	case 113:
		goto st113
	case 468:
		goto st468
	case 469:
		goto st469
	case 114:
		goto st114
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
	case 115:
		goto st115
	case 116:
		goto st116
	case 117:
		goto st117
	case 479:
		goto st479
	case 118:
		goto st118
	case 119:
		goto st119
	case 120:
		goto st120
	case 480:
		goto st480
	case 121:
		goto st121
	case 122:
		goto st122
	case 481:
		goto st481
	case 482:
		goto st482
	case 123:
		goto st123
	case 124:
		goto st124
	case 125:
		goto st125
	case 126:
		goto st126
	case 483:
		goto st483
	case 484:
		goto st484
	case 485:
		goto st485
	case 127:
		goto st127
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
	case 502:
		goto st502
	case 503:
		goto st503
	case 504:
		goto st504
	case 505:
		goto st505
	case 128:
		goto st128
	case 129:
		goto st129
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
	case 512:
		goto st512
	case 513:
		goto st513
	case 514:
		goto st514
	case 130:
		goto st130
	case 131:
		goto st131
	case 132:
		goto st132
	case 515:
		goto st515
	case 133:
		goto st133
	case 134:
		goto st134
	case 135:
		goto st135
	case 516:
		goto st516
	case 136:
		goto st136
	case 137:
		goto st137
	case 517:
		goto st517
	case 518:
		goto st518
	case 138:
		goto st138
	case 139:
		goto st139
	case 140:
		goto st140
	case 519:
		goto st519
	case 520:
		goto st520
	case 141:
		goto st141
	case 521:
		goto st521
	case 142:
		goto st142
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
	case 143:
		goto st143
	case 144:
		goto st144
	case 145:
		goto st145
	case 530:
		goto st530
	case 146:
		goto st146
	case 147:
		goto st147
	case 148:
		goto st148
	case 531:
		goto st531
	case 149:
		goto st149
	case 150:
		goto st150
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
	case 547:
		goto st547
	case 548:
		goto st548
	case 549:
		goto st549
	case 550:
		goto st550
	case 551:
		goto st551
	case 151:
		goto st151
	case 152:
		goto st152
	case 552:
		goto st552
	case 553:
		goto st553
	case 554:
		goto st554
	case 153:
		goto st153
	case 555:
		goto st555
	case 556:
		goto st556
	case 154:
		goto st154
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
	case 562:
		goto st562
	case 563:
		goto st563
	case 564:
		goto st564
	case 565:
		goto st565
	case 566:
		goto st566
	case 567:
		goto st567
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
	case 574:
		goto st574
	case 155:
		goto st155
	case 156:
		goto st156
	case 575:
		goto st575
	case 157:
		goto st157
	case 576:
		goto st576
	case 577:
		goto st577
	case 578:
		goto st578
	case 579:
		goto st579
	case 580:
		goto st580
	case 581:
		goto st581
	case 582:
		goto st582
	case 583:
		goto st583
	case 158:
		goto st158
	case 159:
		goto st159
	case 160:
		goto st160
	case 584:
		goto st584
	case 161:
		goto st161
	case 162:
		goto st162
	case 163:
		goto st163
	case 585:
		goto st585
	case 164:
		goto st164
	case 165:
		goto st165
	case 586:
		goto st586
	case 587:
		goto st587
	case 166:
		goto st166
	case 167:
		goto st167
	case 168:
		goto st168
	case 169:
		goto st169
	case 170:
		goto st170
	case 171:
		goto st171
	case 588:
		goto st588
	case 589:
		goto st589
	case 590:
		goto st590
	case 591:
		goto st591
	case 592:
		goto st592
	case 593:
		goto st593
	case 594:
		goto st594
	case 595:
		goto st595
	case 596:
		goto st596
	case 597:
		goto st597
	case 598:
		goto st598
	case 599:
		goto st599
	case 600:
		goto st600
	case 601:
		goto st601
	case 602:
		goto st602
	case 603:
		goto st603
	case 604:
		goto st604
	case 605:
		goto st605
	case 606:
		goto st606
	case 172:
		goto st172
	case 173:
		goto st173
	case 174:
		goto st174
	case 607:
		goto st607
	case 608:
		goto st608
	case 609:
		goto st609
	case 175:
		goto st175
	case 610:
		goto st610
	case 611:
		goto st611
	case 176:
		goto st176
	case 612:
		goto st612
	case 613:
		goto st613
	case 614:
		goto st614
	case 615:
		goto st615
	case 616:
		goto st616
	case 177:
		goto st177
	case 178:
		goto st178
	case 179:
		goto st179
	case 617:
		goto st617
	case 180:
		goto st180
	case 181:
		goto st181
	case 182:
		goto st182
	case 618:
		goto st618
	case 183:
		goto st183
	case 184:
		goto st184
	case 619:
		goto st619
	case 620:
		goto st620
	case 185:
		goto st185
	case 621:
		goto st621
	case 622:
		goto st622
	case 186:
		goto st186
	case 187:
		goto st187
	case 188:
		goto st188
	case 623:
		goto st623
	case 189:
		goto st189
	case 190:
		goto st190
	case 624:
		goto st624
	case 625:
		goto st625
	case 626:
		goto st626
	case 627:
		goto st627
	case 628:
		goto st628
	case 629:
		goto st629
	case 630:
		goto st630
	case 631:
		goto st631
	case 191:
		goto st191
	case 192:
		goto st192
	case 193:
		goto st193
	case 632:
		goto st632
	case 194:
		goto st194
	case 195:
		goto st195
	case 196:
		goto st196
	case 633:
		goto st633
	case 197:
		goto st197
	case 198:
		goto st198
	case 634:
		goto st634
	case 635:
		goto st635
	case 199:
		goto st199
	case 200:
		goto st200
	case 201:
		goto st201
	case 636:
		goto st636
	case 637:
		goto st637
	case 638:
		goto st638
	case 639:
		goto st639
	case 640:
		goto st640
	case 641:
		goto st641
	case 642:
		goto st642
	case 643:
		goto st643
	case 644:
		goto st644
	case 645:
		goto st645
	case 646:
		goto st646
	case 647:
		goto st647
	case 648:
		goto st648
	case 649:
		goto st649
	case 650:
		goto st650
	case 651:
		goto st651
	case 652:
		goto st652
	case 653:
		goto st653
	case 654:
		goto st654
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
	case 655:
		goto st655
	case 207:
		goto st207
	case 208:
		goto st208
	case 656:
		goto st656
	case 657:
		goto st657
	case 658:
		goto st658
	case 659:
		goto st659
	case 660:
		goto st660
	case 661:
		goto st661
	case 662:
		goto st662
	case 663:
		goto st663
	case 664:
		goto st664
	case 209:
		goto st209
	case 210:
		goto st210
	case 211:
		goto st211
	case 665:
		goto st665
	case 212:
		goto st212
	case 213:
		goto st213
	case 214:
		goto st214
	case 666:
		goto st666
	case 215:
		goto st215
	case 216:
		goto st216
	case 667:
		goto st667
	case 668:
		goto st668
	case 217:
		goto st217
	case 218:
		goto st218
	case 219:
		goto st219
	case 220:
		goto st220
	case 669:
		goto st669
	case 221:
		goto st221
	case 222:
		goto st222
	case 670:
		goto st670
	case 671:
		goto st671
	case 672:
		goto st672
	case 673:
		goto st673
	case 674:
		goto st674
	case 675:
		goto st675
	case 676:
		goto st676
	case 677:
		goto st677
	case 223:
		goto st223
	case 224:
		goto st224
	case 225:
		goto st225
	case 678:
		goto st678
	case 226:
		goto st226
	case 227:
		goto st227
	case 228:
		goto st228
	case 679:
		goto st679
	case 229:
		goto st229
	case 230:
		goto st230
	case 680:
		goto st680
	case 681:
		goto st681
	case 231:
		goto st231
	case 232:
		goto st232
	case 233:
		goto st233
	case 682:
		goto st682
	case 683:
		goto st683
	case 684:
		goto st684
	case 685:
		goto st685
	case 686:
		goto st686
	case 687:
		goto st687
	case 688:
		goto st688
	case 689:
		goto st689
	case 690:
		goto st690
	case 691:
		goto st691
	case 692:
		goto st692
	case 693:
		goto st693
	case 694:
		goto st694
	case 695:
		goto st695
	case 696:
		goto st696
	case 697:
		goto st697
	case 698:
		goto st698
	case 699:
		goto st699
	case 700:
		goto st700
	case 234:
		goto st234
	case 235:
		goto st235
	case 701:
		goto st701
	case 236:
		goto st236
	case 237:
		goto st237
	case 702:
		goto st702
	case 703:
		goto st703
	case 704:
		goto st704
	case 705:
		goto st705
	case 706:
		goto st706
	case 707:
		goto st707
	case 708:
		goto st708
	case 709:
		goto st709
	case 238:
		goto st238
	case 239:
		goto st239
	case 240:
		goto st240
	case 710:
		goto st710
	case 241:
		goto st241
	case 242:
		goto st242
	case 243:
		goto st243
	case 711:
		goto st711
	case 244:
		goto st244
	case 245:
		goto st245
	case 712:
		goto st712
	case 713:
		goto st713
	case 246:
		goto st246
	case 247:
		goto st247
	case 714:
		goto st714
	case 250:
		goto st250
	case 717:
		goto st717
	case 718:
		goto st718
	case 251:
		goto st251
	case 252:
		goto st252
	case 253:
		goto st253
	case 254:
		goto st254
	case 719:
		goto st719
	case 255:
		goto st255
	case 720:
		goto st720
	case 256:
		goto st256
	case 257:
		goto st257
	case 258:
		goto st258
	case 715:
		goto st715
	case 716:
		goto st716
	case 248:
		goto st248
	case 249:
		goto st249
	}

	if ( m.p)++; ( m.p) == ( m.pe) {
		goto _test_eof
	}
_resume:
	switch ( m.cs) {
	case 259:
		goto st_case_259
	case 1:
		goto st_case_1
	case 2:
		goto st_case_2
	case 3:
		goto st_case_3
	case 0:
		goto st_case_0
	case 4:
		goto st_case_4
	case 5:
		goto st_case_5
	case 6:
		goto st_case_6
	case 7:
		goto st_case_7
	case 8:
		goto st_case_8
	case 260:
		goto st_case_260
	case 261:
		goto st_case_261
	case 262:
		goto st_case_262
	case 9:
		goto st_case_9
	case 10:
		goto st_case_10
	case 11:
		goto st_case_11
	case 12:
		goto st_case_12
	case 13:
		goto st_case_13
	case 14:
		goto st_case_14
	case 15:
		goto st_case_15
	case 16:
		goto st_case_16
	case 17:
		goto st_case_17
	case 18:
		goto st_case_18
	case 19:
		goto st_case_19
	case 20:
		goto st_case_20
	case 21:
		goto st_case_21
	case 22:
		goto st_case_22
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
	case 263:
		goto st_case_263
	case 264:
		goto st_case_264
	case 34:
		goto st_case_34
	case 35:
		goto st_case_35
	case 265:
		goto st_case_265
	case 266:
		goto st_case_266
	case 267:
		goto st_case_267
	case 36:
		goto st_case_36
	case 268:
		goto st_case_268
	case 269:
		goto st_case_269
	case 270:
		goto st_case_270
	case 271:
		goto st_case_271
	case 272:
		goto st_case_272
	case 273:
		goto st_case_273
	case 274:
		goto st_case_274
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
	case 282:
		goto st_case_282
	case 283:
		goto st_case_283
	case 284:
		goto st_case_284
	case 285:
		goto st_case_285
	case 37:
		goto st_case_37
	case 38:
		goto st_case_38
	case 286:
		goto st_case_286
	case 287:
		goto st_case_287
	case 288:
		goto st_case_288
	case 39:
		goto st_case_39
	case 40:
		goto st_case_40
	case 41:
		goto st_case_41
	case 42:
		goto st_case_42
	case 43:
		goto st_case_43
	case 289:
		goto st_case_289
	case 290:
		goto st_case_290
	case 291:
		goto st_case_291
	case 292:
		goto st_case_292
	case 44:
		goto st_case_44
	case 293:
		goto st_case_293
	case 294:
		goto st_case_294
	case 295:
		goto st_case_295
	case 296:
		goto st_case_296
	case 297:
		goto st_case_297
	case 298:
		goto st_case_298
	case 299:
		goto st_case_299
	case 300:
		goto st_case_300
	case 301:
		goto st_case_301
	case 302:
		goto st_case_302
	case 303:
		goto st_case_303
	case 304:
		goto st_case_304
	case 305:
		goto st_case_305
	case 306:
		goto st_case_306
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
	case 45:
		goto st_case_45
	case 46:
		goto st_case_46
	case 47:
		goto st_case_47
	case 48:
		goto st_case_48
	case 49:
		goto st_case_49
	case 50:
		goto st_case_50
	case 51:
		goto st_case_51
	case 52:
		goto st_case_52
	case 53:
		goto st_case_53
	case 54:
		goto st_case_54
	case 315:
		goto st_case_315
	case 316:
		goto st_case_316
	case 317:
		goto st_case_317
	case 55:
		goto st_case_55
	case 56:
		goto st_case_56
	case 57:
		goto st_case_57
	case 58:
		goto st_case_58
	case 59:
		goto st_case_59
	case 60:
		goto st_case_60
	case 318:
		goto st_case_318
	case 319:
		goto st_case_319
	case 61:
		goto st_case_61
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
	case 325:
		goto st_case_325
	case 326:
		goto st_case_326
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
	case 62:
		goto st_case_62
	case 340:
		goto st_case_340
	case 341:
		goto st_case_341
	case 342:
		goto st_case_342
	case 63:
		goto st_case_63
	case 343:
		goto st_case_343
	case 344:
		goto st_case_344
	case 345:
		goto st_case_345
	case 346:
		goto st_case_346
	case 347:
		goto st_case_347
	case 348:
		goto st_case_348
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
	case 64:
		goto st_case_64
	case 65:
		goto st_case_65
	case 66:
		goto st_case_66
	case 67:
		goto st_case_67
	case 68:
		goto st_case_68
	case 363:
		goto st_case_363
	case 69:
		goto st_case_69
	case 70:
		goto st_case_70
	case 71:
		goto st_case_71
	case 72:
		goto st_case_72
	case 73:
		goto st_case_73
	case 364:
		goto st_case_364
	case 365:
		goto st_case_365
	case 366:
		goto st_case_366
	case 74:
		goto st_case_74
	case 75:
		goto st_case_75
	case 367:
		goto st_case_367
	case 368:
		goto st_case_368
	case 76:
		goto st_case_76
	case 369:
		goto st_case_369
	case 77:
		goto st_case_77
	case 370:
		goto st_case_370
	case 371:
		goto st_case_371
	case 372:
		goto st_case_372
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
	case 380:
		goto st_case_380
	case 381:
		goto st_case_381
	case 382:
		goto st_case_382
	case 383:
		goto st_case_383
	case 384:
		goto st_case_384
	case 385:
		goto st_case_385
	case 386:
		goto st_case_386
	case 387:
		goto st_case_387
	case 388:
		goto st_case_388
	case 389:
		goto st_case_389
	case 78:
		goto st_case_78
	case 79:
		goto st_case_79
	case 80:
		goto st_case_80
	case 81:
		goto st_case_81
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
	case 390:
		goto st_case_390
	case 391:
		goto st_case_391
	case 392:
		goto st_case_392
	case 393:
		goto st_case_393
	case 92:
		goto st_case_92
	case 93:
		goto st_case_93
	case 94:
		goto st_case_94
	case 95:
		goto st_case_95
	case 394:
		goto st_case_394
	case 395:
		goto st_case_395
	case 96:
		goto st_case_96
	case 97:
		goto st_case_97
	case 396:
		goto st_case_396
	case 98:
		goto st_case_98
	case 99:
		goto st_case_99
	case 397:
		goto st_case_397
	case 398:
		goto st_case_398
	case 100:
		goto st_case_100
	case 399:
		goto st_case_399
	case 400:
		goto st_case_400
	case 101:
		goto st_case_101
	case 102:
		goto st_case_102
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
	case 409:
		goto st_case_409
	case 410:
		goto st_case_410
	case 411:
		goto st_case_411
	case 412:
		goto st_case_412
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
	case 103:
		goto st_case_103
	case 419:
		goto st_case_419
	case 420:
		goto st_case_420
	case 421:
		goto st_case_421
	case 104:
		goto st_case_104
	case 105:
		goto st_case_105
	case 422:
		goto st_case_422
	case 423:
		goto st_case_423
	case 424:
		goto st_case_424
	case 106:
		goto st_case_106
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
	case 433:
		goto st_case_433
	case 434:
		goto st_case_434
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
	case 107:
		goto st_case_107
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
	case 455:
		goto st_case_455
	case 456:
		goto st_case_456
	case 457:
		goto st_case_457
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
	case 108:
		goto st_case_108
	case 109:
		goto st_case_109
	case 110:
		goto st_case_110
	case 111:
		goto st_case_111
	case 112:
		goto st_case_112
	case 467:
		goto st_case_467
	case 113:
		goto st_case_113
	case 468:
		goto st_case_468
	case 469:
		goto st_case_469
	case 114:
		goto st_case_114
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
	case 115:
		goto st_case_115
	case 116:
		goto st_case_116
	case 117:
		goto st_case_117
	case 479:
		goto st_case_479
	case 118:
		goto st_case_118
	case 119:
		goto st_case_119
	case 120:
		goto st_case_120
	case 480:
		goto st_case_480
	case 121:
		goto st_case_121
	case 122:
		goto st_case_122
	case 481:
		goto st_case_481
	case 482:
		goto st_case_482
	case 123:
		goto st_case_123
	case 124:
		goto st_case_124
	case 125:
		goto st_case_125
	case 126:
		goto st_case_126
	case 483:
		goto st_case_483
	case 484:
		goto st_case_484
	case 485:
		goto st_case_485
	case 127:
		goto st_case_127
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
	case 502:
		goto st_case_502
	case 503:
		goto st_case_503
	case 504:
		goto st_case_504
	case 505:
		goto st_case_505
	case 128:
		goto st_case_128
	case 129:
		goto st_case_129
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
	case 512:
		goto st_case_512
	case 513:
		goto st_case_513
	case 514:
		goto st_case_514
	case 130:
		goto st_case_130
	case 131:
		goto st_case_131
	case 132:
		goto st_case_132
	case 515:
		goto st_case_515
	case 133:
		goto st_case_133
	case 134:
		goto st_case_134
	case 135:
		goto st_case_135
	case 516:
		goto st_case_516
	case 136:
		goto st_case_136
	case 137:
		goto st_case_137
	case 517:
		goto st_case_517
	case 518:
		goto st_case_518
	case 138:
		goto st_case_138
	case 139:
		goto st_case_139
	case 140:
		goto st_case_140
	case 519:
		goto st_case_519
	case 520:
		goto st_case_520
	case 141:
		goto st_case_141
	case 521:
		goto st_case_521
	case 142:
		goto st_case_142
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
	case 143:
		goto st_case_143
	case 144:
		goto st_case_144
	case 145:
		goto st_case_145
	case 530:
		goto st_case_530
	case 146:
		goto st_case_146
	case 147:
		goto st_case_147
	case 148:
		goto st_case_148
	case 531:
		goto st_case_531
	case 149:
		goto st_case_149
	case 150:
		goto st_case_150
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
	case 547:
		goto st_case_547
	case 548:
		goto st_case_548
	case 549:
		goto st_case_549
	case 550:
		goto st_case_550
	case 551:
		goto st_case_551
	case 151:
		goto st_case_151
	case 152:
		goto st_case_152
	case 552:
		goto st_case_552
	case 553:
		goto st_case_553
	case 554:
		goto st_case_554
	case 153:
		goto st_case_153
	case 555:
		goto st_case_555
	case 556:
		goto st_case_556
	case 154:
		goto st_case_154
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
	case 562:
		goto st_case_562
	case 563:
		goto st_case_563
	case 564:
		goto st_case_564
	case 565:
		goto st_case_565
	case 566:
		goto st_case_566
	case 567:
		goto st_case_567
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
	case 574:
		goto st_case_574
	case 155:
		goto st_case_155
	case 156:
		goto st_case_156
	case 575:
		goto st_case_575
	case 157:
		goto st_case_157
	case 576:
		goto st_case_576
	case 577:
		goto st_case_577
	case 578:
		goto st_case_578
	case 579:
		goto st_case_579
	case 580:
		goto st_case_580
	case 581:
		goto st_case_581
	case 582:
		goto st_case_582
	case 583:
		goto st_case_583
	case 158:
		goto st_case_158
	case 159:
		goto st_case_159
	case 160:
		goto st_case_160
	case 584:
		goto st_case_584
	case 161:
		goto st_case_161
	case 162:
		goto st_case_162
	case 163:
		goto st_case_163
	case 585:
		goto st_case_585
	case 164:
		goto st_case_164
	case 165:
		goto st_case_165
	case 586:
		goto st_case_586
	case 587:
		goto st_case_587
	case 166:
		goto st_case_166
	case 167:
		goto st_case_167
	case 168:
		goto st_case_168
	case 169:
		goto st_case_169
	case 170:
		goto st_case_170
	case 171:
		goto st_case_171
	case 588:
		goto st_case_588
	case 589:
		goto st_case_589
	case 590:
		goto st_case_590
	case 591:
		goto st_case_591
	case 592:
		goto st_case_592
	case 593:
		goto st_case_593
	case 594:
		goto st_case_594
	case 595:
		goto st_case_595
	case 596:
		goto st_case_596
	case 597:
		goto st_case_597
	case 598:
		goto st_case_598
	case 599:
		goto st_case_599
	case 600:
		goto st_case_600
	case 601:
		goto st_case_601
	case 602:
		goto st_case_602
	case 603:
		goto st_case_603
	case 604:
		goto st_case_604
	case 605:
		goto st_case_605
	case 606:
		goto st_case_606
	case 172:
		goto st_case_172
	case 173:
		goto st_case_173
	case 174:
		goto st_case_174
	case 607:
		goto st_case_607
	case 608:
		goto st_case_608
	case 609:
		goto st_case_609
	case 175:
		goto st_case_175
	case 610:
		goto st_case_610
	case 611:
		goto st_case_611
	case 176:
		goto st_case_176
	case 612:
		goto st_case_612
	case 613:
		goto st_case_613
	case 614:
		goto st_case_614
	case 615:
		goto st_case_615
	case 616:
		goto st_case_616
	case 177:
		goto st_case_177
	case 178:
		goto st_case_178
	case 179:
		goto st_case_179
	case 617:
		goto st_case_617
	case 180:
		goto st_case_180
	case 181:
		goto st_case_181
	case 182:
		goto st_case_182
	case 618:
		goto st_case_618
	case 183:
		goto st_case_183
	case 184:
		goto st_case_184
	case 619:
		goto st_case_619
	case 620:
		goto st_case_620
	case 185:
		goto st_case_185
	case 621:
		goto st_case_621
	case 622:
		goto st_case_622
	case 186:
		goto st_case_186
	case 187:
		goto st_case_187
	case 188:
		goto st_case_188
	case 623:
		goto st_case_623
	case 189:
		goto st_case_189
	case 190:
		goto st_case_190
	case 624:
		goto st_case_624
	case 625:
		goto st_case_625
	case 626:
		goto st_case_626
	case 627:
		goto st_case_627
	case 628:
		goto st_case_628
	case 629:
		goto st_case_629
	case 630:
		goto st_case_630
	case 631:
		goto st_case_631
	case 191:
		goto st_case_191
	case 192:
		goto st_case_192
	case 193:
		goto st_case_193
	case 632:
		goto st_case_632
	case 194:
		goto st_case_194
	case 195:
		goto st_case_195
	case 196:
		goto st_case_196
	case 633:
		goto st_case_633
	case 197:
		goto st_case_197
	case 198:
		goto st_case_198
	case 634:
		goto st_case_634
	case 635:
		goto st_case_635
	case 199:
		goto st_case_199
	case 200:
		goto st_case_200
	case 201:
		goto st_case_201
	case 636:
		goto st_case_636
	case 637:
		goto st_case_637
	case 638:
		goto st_case_638
	case 639:
		goto st_case_639
	case 640:
		goto st_case_640
	case 641:
		goto st_case_641
	case 642:
		goto st_case_642
	case 643:
		goto st_case_643
	case 644:
		goto st_case_644
	case 645:
		goto st_case_645
	case 646:
		goto st_case_646
	case 647:
		goto st_case_647
	case 648:
		goto st_case_648
	case 649:
		goto st_case_649
	case 650:
		goto st_case_650
	case 651:
		goto st_case_651
	case 652:
		goto st_case_652
	case 653:
		goto st_case_653
	case 654:
		goto st_case_654
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
	case 655:
		goto st_case_655
	case 207:
		goto st_case_207
	case 208:
		goto st_case_208
	case 656:
		goto st_case_656
	case 657:
		goto st_case_657
	case 658:
		goto st_case_658
	case 659:
		goto st_case_659
	case 660:
		goto st_case_660
	case 661:
		goto st_case_661
	case 662:
		goto st_case_662
	case 663:
		goto st_case_663
	case 664:
		goto st_case_664
	case 209:
		goto st_case_209
	case 210:
		goto st_case_210
	case 211:
		goto st_case_211
	case 665:
		goto st_case_665
	case 212:
		goto st_case_212
	case 213:
		goto st_case_213
	case 214:
		goto st_case_214
	case 666:
		goto st_case_666
	case 215:
		goto st_case_215
	case 216:
		goto st_case_216
	case 667:
		goto st_case_667
	case 668:
		goto st_case_668
	case 217:
		goto st_case_217
	case 218:
		goto st_case_218
	case 219:
		goto st_case_219
	case 220:
		goto st_case_220
	case 669:
		goto st_case_669
	case 221:
		goto st_case_221
	case 222:
		goto st_case_222
	case 670:
		goto st_case_670
	case 671:
		goto st_case_671
	case 672:
		goto st_case_672
	case 673:
		goto st_case_673
	case 674:
		goto st_case_674
	case 675:
		goto st_case_675
	case 676:
		goto st_case_676
	case 677:
		goto st_case_677
	case 223:
		goto st_case_223
	case 224:
		goto st_case_224
	case 225:
		goto st_case_225
	case 678:
		goto st_case_678
	case 226:
		goto st_case_226
	case 227:
		goto st_case_227
	case 228:
		goto st_case_228
	case 679:
		goto st_case_679
	case 229:
		goto st_case_229
	case 230:
		goto st_case_230
	case 680:
		goto st_case_680
	case 681:
		goto st_case_681
	case 231:
		goto st_case_231
	case 232:
		goto st_case_232
	case 233:
		goto st_case_233
	case 682:
		goto st_case_682
	case 683:
		goto st_case_683
	case 684:
		goto st_case_684
	case 685:
		goto st_case_685
	case 686:
		goto st_case_686
	case 687:
		goto st_case_687
	case 688:
		goto st_case_688
	case 689:
		goto st_case_689
	case 690:
		goto st_case_690
	case 691:
		goto st_case_691
	case 692:
		goto st_case_692
	case 693:
		goto st_case_693
	case 694:
		goto st_case_694
	case 695:
		goto st_case_695
	case 696:
		goto st_case_696
	case 697:
		goto st_case_697
	case 698:
		goto st_case_698
	case 699:
		goto st_case_699
	case 700:
		goto st_case_700
	case 234:
		goto st_case_234
	case 235:
		goto st_case_235
	case 701:
		goto st_case_701
	case 236:
		goto st_case_236
	case 237:
		goto st_case_237
	case 702:
		goto st_case_702
	case 703:
		goto st_case_703
	case 704:
		goto st_case_704
	case 705:
		goto st_case_705
	case 706:
		goto st_case_706
	case 707:
		goto st_case_707
	case 708:
		goto st_case_708
	case 709:
		goto st_case_709
	case 238:
		goto st_case_238
	case 239:
		goto st_case_239
	case 240:
		goto st_case_240
	case 710:
		goto st_case_710
	case 241:
		goto st_case_241
	case 242:
		goto st_case_242
	case 243:
		goto st_case_243
	case 711:
		goto st_case_711
	case 244:
		goto st_case_244
	case 245:
		goto st_case_245
	case 712:
		goto st_case_712
	case 713:
		goto st_case_713
	case 246:
		goto st_case_246
	case 247:
		goto st_case_247
	case 714:
		goto st_case_714
	case 250:
		goto st_case_250
	case 717:
		goto st_case_717
	case 718:
		goto st_case_718
	case 251:
		goto st_case_251
	case 252:
		goto st_case_252
	case 253:
		goto st_case_253
	case 254:
		goto st_case_254
	case 719:
		goto st_case_719
	case 255:
		goto st_case_255
	case 720:
		goto st_case_720
	case 256:
		goto st_case_256
	case 257:
		goto st_case_257
	case 258:
		goto st_case_258
	case 715:
		goto st_case_715
	case 716:
		goto st_case_716
	case 248:
		goto st_case_248
	case 249:
		goto st_case_249
	}
	goto st_out
	st259:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof259
		}
	st_case_259:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr35
		case 11:
			goto tr440
		case 13:
			goto tr35
		case 32:
			goto tr439
		case 35:
			goto tr35
		case 44:
			goto tr35
		case 92:
			goto tr441
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr439
		}
		goto tr438
tr33:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st1
tr438:
//line plugins/parsers/influx/machine.go.rl:73

	foundMetric = true

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st1
	st1:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof1
		}
	st_case_1:
//line plugins/parsers/influx/machine.go:3096
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr2
		case 11:
			goto tr3
		case 13:
			goto tr2
		case 32:
			goto tr1
		case 44:
			goto tr4
		case 92:
			goto st96
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr1
		}
		goto st1
tr1:
	( m.cs) = 2
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr60:
	( m.cs) = 2
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st2:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof2
		}
	st_case_2:
//line plugins/parsers/influx/machine.go:3146
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr8
		case 11:
			goto tr9
		case 13:
			goto tr8
		case 32:
			goto st2
		case 44:
			goto tr8
		case 61:
			goto tr8
		case 92:
			goto tr10
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st2
		}
		goto tr6
tr6:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st3
	st3:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof3
		}
	st_case_3:
//line plugins/parsers/influx/machine.go:3178
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr8
		case 44:
			goto tr8
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr8
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr8
		}
		goto st3
tr2:
	( m.cs) = 0
//line plugins/parsers/influx/machine.go.rl:37

	err = ErrTagParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; goto _out }

	goto _again
tr8:
	( m.cs) = 0
//line plugins/parsers/influx/machine.go.rl:30

	err = ErrFieldParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; goto _out }

	goto _again
tr35:
	( m.cs) = 0
//line plugins/parsers/influx/machine.go.rl:23

	err = ErrNameParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; goto _out }

	goto _again
tr39:
	( m.cs) = 0
//line plugins/parsers/influx/machine.go.rl:23

	err = ErrNameParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:37

	err = ErrTagParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; goto _out }

	goto _again
tr43:
	( m.cs) = 0
//line plugins/parsers/influx/machine.go.rl:23

	err = ErrNameParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:30

	err = ErrFieldParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; goto _out }

	goto _again
tr47:
	( m.cs) = 0
//line plugins/parsers/influx/machine.go.rl:37

	err = ErrTagParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:30

	err = ErrFieldParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; goto _out }

	goto _again
tr105:
	( m.cs) = 0
//line plugins/parsers/influx/machine.go.rl:30

	err = ErrFieldParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:44

	err = ErrTimestampParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; goto _out }

	goto _again
tr132:
	( m.cs) = 0
//line plugins/parsers/influx/machine.go.rl:37

	err = ErrTagParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:30

	err = ErrFieldParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:44

	err = ErrTimestampParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; goto _out }

	goto _again
tr198:
	( m.cs) = 0
//line plugins/parsers/influx/machine.go.rl:37

	err = ErrTagParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:44

	err = ErrTimestampParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; goto _out }

	goto _again
tr404:
	( m.cs) = 0
//line plugins/parsers/influx/machine.go.rl:23

	err = ErrNameParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:37

	err = ErrTagParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; goto _out }

//line plugins/parsers/influx/machine.go.rl:30

	err = ErrFieldParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; goto _out }

	goto _again
tr407:
	( m.cs) = 0
//line plugins/parsers/influx/machine.go.rl:44

	err = ErrTimestampParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; goto _out }

	goto _again
tr1023:
//line plugins/parsers/influx/machine.go.rl:64

	( m.p)--

	{goto st259 }

	goto st0
//line plugins/parsers/influx/machine.go:3399
st_case_0:
	st0:
		( m.cs) = 0
		goto _out
tr12:
//line plugins/parsers/influx/machine.go.rl:99

	key = m.text()

	goto st4
	st4:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof4
		}
	st_case_4:
//line plugins/parsers/influx/machine.go:3415
		switch ( m.data)[( m.p)] {
		case 34:
			goto st5
		case 45:
			goto tr15
		case 46:
			goto tr16
		case 48:
			goto tr17
		case 70:
			goto tr19
		case 84:
			goto tr20
		case 102:
			goto tr21
		case 116:
			goto tr22
		}
		if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr18
		}
		goto tr8
	st5:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof5
		}
	st_case_5:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr24
		case 12:
			goto tr8
		case 13:
			goto tr25
		case 34:
			goto tr26
		case 92:
			goto tr27
		}
		goto tr23
tr23:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st6
	st6:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof6
		}
	st_case_6:
//line plugins/parsers/influx/machine.go:3467
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		goto st6
tr24:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st7
	st7:
//line plugins/parsers/influx/machine.go.rl:157

	m.lineno++
	m.sol = m.p
	m.sol++ // next char will be the first column in the line

		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof7
		}
	st_case_7:
//line plugins/parsers/influx/machine.go:3498
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		goto st6
tr25:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st8
	st8:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof8
		}
	st_case_8:
//line plugins/parsers/influx/machine.go:3523
		if ( m.data)[( m.p)] == 10 {
			goto st7
		}
		goto tr8
tr26:
	( m.cs) = 260
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr31:
	( m.cs) = 260
//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st260:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof260
		}
	st_case_260:
//line plugins/parsers/influx/machine.go:3563
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 13:
			goto st34
		case 32:
			goto st261
		case 44:
			goto st37
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st261
		}
		goto tr105
tr516:
	( m.cs) = 261
//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr909:
	( m.cs) = 261
//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr912:
	( m.cs) = 261
//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr916:
	( m.cs) = 261
//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st261:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof261
		}
	st_case_261:
//line plugins/parsers/influx/machine.go:3635
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 13:
			goto st34
		case 32:
			goto st261
		case 45:
			goto tr445
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr446
			}
		case ( m.data)[( m.p)] >= 9:
			goto st261
		}
		goto tr407
tr451:
	( m.cs) = 262
//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr715:
	( m.cs) = 262
//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr925:
	( m.cs) = 262
//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr930:
	( m.cs) = 262
//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr935:
	( m.cs) = 262
//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st262:
//line plugins/parsers/influx/machine.go.rl:157

	m.lineno++
	m.sol = m.p
	m.sol++ // next char will be the first column in the line

//line plugins/parsers/influx/machine.go.rl:163

	( m.cs) = 715;
	{( m.p)++; goto _out }

		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof262
		}
	st_case_262:
//line plugins/parsers/influx/machine.go:3736
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr35
		case 11:
			goto tr36
		case 13:
			goto tr35
		case 32:
			goto st9
		case 35:
			goto tr35
		case 44:
			goto tr35
		case 92:
			goto tr37
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st9
		}
		goto tr33
tr439:
//line plugins/parsers/influx/machine.go.rl:73

	foundMetric = true

	goto st9
	st9:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof9
		}
	st_case_9:
//line plugins/parsers/influx/machine.go:3768
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr35
		case 11:
			goto tr36
		case 13:
			goto tr35
		case 32:
			goto st9
		case 35:
			goto tr35
		case 44:
			goto tr35
		case 92:
			goto tr37
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st9
		}
		goto tr33
tr36:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st10
tr440:
//line plugins/parsers/influx/machine.go.rl:73

	foundMetric = true

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st10
	st10:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof10
		}
	st_case_10:
//line plugins/parsers/influx/machine.go:3810
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr39
		case 11:
			goto tr40
		case 13:
			goto tr39
		case 32:
			goto tr38
		case 35:
			goto st1
		case 44:
			goto tr4
		case 92:
			goto tr37
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr38
		}
		goto tr33
tr38:
	( m.cs) = 11
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st11:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof11
		}
	st_case_11:
//line plugins/parsers/influx/machine.go:3849
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr43
		case 11:
			goto tr44
		case 13:
			goto tr43
		case 32:
			goto st11
		case 35:
			goto tr6
		case 44:
			goto tr43
		case 61:
			goto tr33
		case 92:
			goto tr45
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st11
		}
		goto tr41
tr41:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st12
	st12:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof12
		}
	st_case_12:
//line plugins/parsers/influx/machine.go:3883
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr48
		case 13:
			goto tr47
		case 32:
			goto tr1
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr1
		}
		goto st12
tr48:
	( m.cs) = 13
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr51:
	( m.cs) = 13
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
	st13:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof13
		}
	st_case_13:
//line plugins/parsers/influx/machine.go:3939
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr51
		case 13:
			goto tr47
		case 32:
			goto tr1
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto tr45
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr1
		}
		goto tr41
tr4:
	( m.cs) = 14
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr62:
	( m.cs) = 14
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st14:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof14
		}
	st_case_14:
//line plugins/parsers/influx/machine.go:3991
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr2
		case 44:
			goto tr2
		case 61:
			goto tr2
		case 92:
			goto tr53
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr2
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr2
		}
		goto tr52
tr52:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st15
	st15:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof15
		}
	st_case_15:
//line plugins/parsers/influx/machine.go:4022
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr2
		case 44:
			goto tr2
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr2
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr2
		}
		goto st15
tr55:
//line plugins/parsers/influx/machine.go.rl:86

	key = m.text()

	goto st16
	st16:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof16
		}
	st_case_16:
//line plugins/parsers/influx/machine.go:4053
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr2
		case 44:
			goto tr2
		case 61:
			goto tr2
		case 92:
			goto tr58
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr2
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr2
		}
		goto tr57
tr57:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st17
	st17:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof17
		}
	st_case_17:
//line plugins/parsers/influx/machine.go:4084
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr2
		case 11:
			goto tr61
		case 13:
			goto tr2
		case 32:
			goto tr60
		case 44:
			goto tr62
		case 61:
			goto tr2
		case 92:
			goto st23
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr60
		}
		goto st17
tr61:
	( m.cs) = 18
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st18:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof18
		}
	st_case_18:
//line plugins/parsers/influx/machine.go:4123
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr65
		case 13:
			goto tr47
		case 32:
			goto tr60
		case 44:
			goto tr62
		case 61:
			goto tr47
		case 92:
			goto tr66
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr60
		}
		goto tr64
tr64:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st19
	st19:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof19
		}
	st_case_19:
//line plugins/parsers/influx/machine.go:4155
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr68
		case 13:
			goto tr47
		case 32:
			goto tr60
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr60
		}
		goto st19
tr68:
	( m.cs) = 20
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr65:
	( m.cs) = 20
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
	st20:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof20
		}
	st_case_20:
//line plugins/parsers/influx/machine.go:4211
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr65
		case 13:
			goto tr47
		case 32:
			goto tr60
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto tr66
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr60
		}
		goto tr64
tr66:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st21
	st21:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof21
		}
	st_case_21:
//line plugins/parsers/influx/machine.go:4243
		if ( m.data)[( m.p)] == 92 {
			goto st22
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr47
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr47
		}
		goto st19
	st22:
//line plugins/parsers/influx/machine.go.rl:234
 ( m.p)--
 
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof22
		}
	st_case_22:
//line plugins/parsers/influx/machine.go:4264
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr68
		case 13:
			goto tr47
		case 32:
			goto tr60
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr60
		}
		goto st19
tr58:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st23
	st23:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof23
		}
	st_case_23:
//line plugins/parsers/influx/machine.go:4296
		if ( m.data)[( m.p)] == 92 {
			goto st24
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr2
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr2
		}
		goto st17
	st24:
//line plugins/parsers/influx/machine.go.rl:234
 ( m.p)--
 
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof24
		}
	st_case_24:
//line plugins/parsers/influx/machine.go:4317
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr2
		case 11:
			goto tr61
		case 13:
			goto tr2
		case 32:
			goto tr60
		case 44:
			goto tr62
		case 61:
			goto tr2
		case 92:
			goto st23
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr60
		}
		goto st17
tr53:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st25
	st25:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof25
		}
	st_case_25:
//line plugins/parsers/influx/machine.go:4349
		if ( m.data)[( m.p)] == 92 {
			goto st26
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr2
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr2
		}
		goto st15
	st26:
//line plugins/parsers/influx/machine.go.rl:234
 ( m.p)--
 
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof26
		}
	st_case_26:
//line plugins/parsers/influx/machine.go:4370
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr2
		case 44:
			goto tr2
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr2
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr2
		}
		goto st15
tr49:
//line plugins/parsers/influx/machine.go.rl:99

	key = m.text()

	goto st27
tr406:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:99

	key = m.text()

	goto st27
	st27:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof27
		}
	st_case_27:
//line plugins/parsers/influx/machine.go:4411
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr3
		case 13:
			goto tr47
		case 32:
			goto tr1
		case 34:
			goto st30
		case 44:
			goto tr4
		case 45:
			goto tr74
		case 46:
			goto tr75
		case 48:
			goto tr76
		case 70:
			goto tr78
		case 84:
			goto tr79
		case 92:
			goto st96
		case 102:
			goto tr80
		case 116:
			goto tr81
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr77
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr1
		}
		goto st1
tr3:
	( m.cs) = 28
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st28:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof28
		}
	st_case_28:
//line plugins/parsers/influx/machine.go:4469
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr51
		case 13:
			goto tr47
		case 32:
			goto tr1
		case 44:
			goto tr4
		case 61:
			goto st1
		case 92:
			goto tr45
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr1
		}
		goto tr41
tr45:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st29
	st29:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof29
		}
	st_case_29:
//line plugins/parsers/influx/machine.go:4501
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr8
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr8
		}
		goto st12
	st30:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof30
		}
	st_case_30:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr83
		case 10:
			goto tr24
		case 11:
			goto tr84
		case 12:
			goto tr1
		case 13:
			goto tr25
		case 32:
			goto tr83
		case 34:
			goto tr85
		case 44:
			goto tr86
		case 92:
			goto tr87
		}
		goto tr82
tr82:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st31
	st31:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof31
		}
	st_case_31:
//line plugins/parsers/influx/machine.go:4548
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr89
		case 10:
			goto st7
		case 11:
			goto tr90
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr89
		case 34:
			goto tr91
		case 44:
			goto tr92
		case 92:
			goto st142
		}
		goto st31
tr89:
	( m.cs) = 32
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr83:
	( m.cs) = 32
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
tr231:
	( m.cs) = 32
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st32:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof32
		}
	st_case_32:
//line plugins/parsers/influx/machine.go:4618
		switch ( m.data)[( m.p)] {
		case 9:
			goto st32
		case 10:
			goto st7
		case 11:
			goto tr96
		case 12:
			goto st2
		case 13:
			goto st8
		case 32:
			goto st32
		case 34:
			goto tr97
		case 44:
			goto st6
		case 61:
			goto st6
		case 92:
			goto tr98
		}
		goto tr94
tr94:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st33
	st33:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof33
		}
	st_case_33:
//line plugins/parsers/influx/machine.go:4653
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		goto st33
tr97:
	( m.cs) = 263
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr100:
	( m.cs) = 263
//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr377:
	( m.cs) = 263
//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
	st263:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof263
		}
	st_case_263:
//line plugins/parsers/influx/machine.go:4727
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto st264
		case 13:
			goto st34
		case 32:
			goto st261
		case 44:
			goto st37
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st261
		}
		goto st3
	st264:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof264
		}
	st_case_264:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto st264
		case 13:
			goto st34
		case 32:
			goto st261
		case 44:
			goto tr105
		case 45:
			goto tr448
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr449
			}
		case ( m.data)[( m.p)] >= 9:
			goto st261
		}
		goto st3
tr453:
	( m.cs) = 34
//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr717:
	( m.cs) = 34
//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr927:
	( m.cs) = 34
//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr932:
	( m.cs) = 34
//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr937:
	( m.cs) = 34
//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st34:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof34
		}
	st_case_34:
//line plugins/parsers/influx/machine.go:4850
		if ( m.data)[( m.p)] == 10 {
			goto st262
		}
		goto st0
tr448:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st35
	st35:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof35
		}
	st_case_35:
//line plugins/parsers/influx/machine.go:4866
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr105
		case 44:
			goto tr105
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] < 12:
			if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 10 {
				goto tr105
			}
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st265
			}
		default:
			goto tr105
		}
		goto st3
tr449:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st265
	st265:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof265
		}
	st_case_265:
//line plugins/parsers/influx/machine.go:4901
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr452
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr105
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st268
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st3
tr450:
	( m.cs) = 266
//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st266:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof266
		}
	st_case_266:
//line plugins/parsers/influx/machine.go:4945
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 13:
			goto st34
		case 32:
			goto st266
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st266
		}
		goto st0
tr452:
	( m.cs) = 267
//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st267:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof267
		}
	st_case_267:
//line plugins/parsers/influx/machine.go:4976
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto st267
		case 13:
			goto st34
		case 32:
			goto st266
		case 44:
			goto tr8
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st266
		}
		goto st3
tr10:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st36
	st36:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof36
		}
	st_case_36:
//line plugins/parsers/influx/machine.go:5008
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr8
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr8
		}
		goto st3
	st268:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof268
		}
	st_case_268:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr452
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr105
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st269
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st3
	st269:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof269
		}
	st_case_269:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr452
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr105
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st270
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st3
	st270:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof270
		}
	st_case_270:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr452
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr105
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st271
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st3
	st271:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof271
		}
	st_case_271:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr452
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr105
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st272
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st3
	st272:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof272
		}
	st_case_272:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr452
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr105
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st273
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st3
	st273:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof273
		}
	st_case_273:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr452
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr105
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st274
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st3
	st274:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof274
		}
	st_case_274:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr452
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr105
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st275
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st3
	st275:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof275
		}
	st_case_275:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr452
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr105
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st276
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st3
	st276:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof276
		}
	st_case_276:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr452
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr105
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st277
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st3
	st277:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof277
		}
	st_case_277:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr452
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr105
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st278
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st3
	st278:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof278
		}
	st_case_278:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr452
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr105
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st279
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st3
	st279:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof279
		}
	st_case_279:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr452
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr105
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st280
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st3
	st280:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof280
		}
	st_case_280:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr452
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr105
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st281
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st3
	st281:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof281
		}
	st_case_281:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr452
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr105
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st282
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st3
	st282:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof282
		}
	st_case_282:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr452
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr105
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st283
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st3
	st283:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof283
		}
	st_case_283:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr452
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr105
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st284
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st3
	st284:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof284
		}
	st_case_284:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr452
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr105
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st285
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st3
	st285:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof285
		}
	st_case_285:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr452
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr105
		case 61:
			goto tr12
		case 92:
			goto st36
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr450
		}
		goto st3
tr907:
	( m.cs) = 37
//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr1014:
	( m.cs) = 37
//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr1016:
	( m.cs) = 37
//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr1018:
	( m.cs) = 37
//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st37:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof37
		}
	st_case_37:
//line plugins/parsers/influx/machine.go:5610
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr8
		case 44:
			goto tr8
		case 61:
			goto tr8
		case 92:
			goto tr10
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr8
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr8
		}
		goto tr6
tr101:
//line plugins/parsers/influx/machine.go.rl:99

	key = m.text()

	goto st38
	st38:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof38
		}
	st_case_38:
//line plugins/parsers/influx/machine.go:5641
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr107
		case 45:
			goto tr108
		case 46:
			goto tr109
		case 48:
			goto tr110
		case 70:
			goto tr112
		case 84:
			goto tr113
		case 92:
			goto st76
		case 102:
			goto tr114
		case 116:
			goto tr115
		}
		if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr111
		}
		goto st6
tr107:
	( m.cs) = 286
//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st286:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof286
		}
	st_case_286:
//line plugins/parsers/influx/machine.go:5690
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr475
		case 12:
			goto st261
		case 13:
			goto tr476
		case 32:
			goto tr474
		case 34:
			goto tr26
		case 44:
			goto tr477
		case 92:
			goto tr27
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr474
		}
		goto tr23
tr474:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st287
tr961:
	( m.cs) = 287
//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr966:
	( m.cs) = 287
//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr969:
	( m.cs) = 287
//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr972:
	( m.cs) = 287
//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st287:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof287
		}
	st_case_287:
//line plugins/parsers/influx/machine.go:5774
		switch ( m.data)[( m.p)] {
		case 10:
			goto st288
		case 12:
			goto st261
		case 13:
			goto st74
		case 32:
			goto st287
		case 34:
			goto tr31
		case 45:
			goto tr480
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr481
			}
		case ( m.data)[( m.p)] >= 9:
			goto st287
		}
		goto st6
tr475:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st288
tr584:
	( m.cs) = 288
//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr620:
	( m.cs) = 288
//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr778:
	( m.cs) = 288
//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr784:
	( m.cs) = 288
//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr790:
	( m.cs) = 288
//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st288:
//line plugins/parsers/influx/machine.go.rl:157

	m.lineno++
	m.sol = m.p
	m.sol++ // next char will be the first column in the line

//line plugins/parsers/influx/machine.go.rl:163

	( m.cs) = 715;
	{( m.p)++; goto _out }

		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof288
		}
	st_case_288:
//line plugins/parsers/influx/machine.go:5887
		switch ( m.data)[( m.p)] {
		case 9:
			goto st39
		case 10:
			goto st7
		case 11:
			goto tr117
		case 12:
			goto st9
		case 13:
			goto st8
		case 32:
			goto st39
		case 34:
			goto tr118
		case 35:
			goto st6
		case 44:
			goto st6
		case 92:
			goto tr87
		}
		goto tr82
	st39:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof39
		}
	st_case_39:
		switch ( m.data)[( m.p)] {
		case 9:
			goto st39
		case 10:
			goto st7
		case 11:
			goto tr117
		case 12:
			goto st9
		case 13:
			goto st8
		case 32:
			goto st39
		case 34:
			goto tr118
		case 35:
			goto st6
		case 44:
			goto st6
		case 92:
			goto tr87
		}
		goto tr82
tr117:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st40
	st40:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof40
		}
	st_case_40:
//line plugins/parsers/influx/machine.go:5950
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr119
		case 10:
			goto st7
		case 11:
			goto tr120
		case 12:
			goto tr38
		case 13:
			goto st8
		case 32:
			goto tr119
		case 34:
			goto tr85
		case 35:
			goto st31
		case 44:
			goto tr92
		case 92:
			goto tr87
		}
		goto tr82
tr119:
	( m.cs) = 41
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st41:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof41
		}
	st_case_41:
//line plugins/parsers/influx/machine.go:5992
		switch ( m.data)[( m.p)] {
		case 9:
			goto st41
		case 10:
			goto st7
		case 11:
			goto tr123
		case 12:
			goto st11
		case 13:
			goto st8
		case 32:
			goto st41
		case 34:
			goto tr124
		case 35:
			goto tr94
		case 44:
			goto st6
		case 61:
			goto tr82
		case 92:
			goto tr125
		}
		goto tr121
tr121:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st42
	st42:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof42
		}
	st_case_42:
//line plugins/parsers/influx/machine.go:6029
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr89
		case 10:
			goto st7
		case 11:
			goto tr127
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr89
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		goto st42
tr127:
	( m.cs) = 43
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr131:
	( m.cs) = 43
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
	st43:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof43
		}
	st_case_43:
//line plugins/parsers/influx/machine.go:6088
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr89
		case 10:
			goto st7
		case 11:
			goto tr131
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr89
		case 34:
			goto tr124
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto tr125
		}
		goto tr121
tr124:
	( m.cs) = 289
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr128:
	( m.cs) = 289
//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st289:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof289
		}
	st_case_289:
//line plugins/parsers/influx/machine.go:6147
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto tr483
		case 13:
			goto st34
		case 32:
			goto tr482
		case 44:
			goto tr484
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr482
		}
		goto st12
tr482:
	( m.cs) = 290
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr547:
	( m.cs) = 290
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr622:
	( m.cs) = 290
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr712:
	( m.cs) = 290
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr724:
	( m.cs) = 290
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr731:
	( m.cs) = 290
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr738:
	( m.cs) = 290
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr804:
	( m.cs) = 290
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr809:
	( m.cs) = 290
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr814:
	( m.cs) = 290
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st290:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof290
		}
	st_case_290:
//line plugins/parsers/influx/machine.go:6383
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto tr486
		case 13:
			goto st34
		case 32:
			goto st290
		case 44:
			goto tr105
		case 45:
			goto tr448
		case 61:
			goto tr105
		case 92:
			goto tr10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr449
			}
		case ( m.data)[( m.p)] >= 9:
			goto st290
		}
		goto tr6
tr486:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st291
	st291:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof291
		}
	st_case_291:
//line plugins/parsers/influx/machine.go:6422
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto tr486
		case 13:
			goto st34
		case 32:
			goto st290
		case 44:
			goto tr105
		case 45:
			goto tr448
		case 61:
			goto tr12
		case 92:
			goto tr10
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr449
			}
		case ( m.data)[( m.p)] >= 9:
			goto st290
		}
		goto tr6
tr483:
	( m.cs) = 292
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr487:
	( m.cs) = 292
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
	st292:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof292
		}
	st_case_292:
//line plugins/parsers/influx/machine.go:6485
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto tr487
		case 13:
			goto st34
		case 32:
			goto tr482
		case 44:
			goto tr4
		case 45:
			goto tr488
		case 61:
			goto tr49
		case 92:
			goto tr45
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr489
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr482
		}
		goto tr41
tr488:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st44
	st44:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof44
		}
	st_case_44:
//line plugins/parsers/influx/machine.go:6524
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr132
		case 11:
			goto tr48
		case 13:
			goto tr132
		case 32:
			goto tr1
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st293
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr1
		}
		goto st12
tr489:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st293
	st293:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof293
		}
	st_case_293:
//line plugins/parsers/influx/machine.go:6561
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr491
		case 13:
			goto tr453
		case 32:
			goto tr490
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st297
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr490
		}
		goto st12
tr495:
	( m.cs) = 294
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr556:
	( m.cs) = 294
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr490:
	( m.cs) = 294
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr553:
	( m.cs) = 294
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st294:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof294
		}
	st_case_294:
//line plugins/parsers/influx/machine.go:6664
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto tr494
		case 13:
			goto st34
		case 32:
			goto st294
		case 44:
			goto tr8
		case 61:
			goto tr8
		case 92:
			goto tr10
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st294
		}
		goto tr6
tr494:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st295
	st295:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof295
		}
	st_case_295:
//line plugins/parsers/influx/machine.go:6696
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto tr494
		case 13:
			goto st34
		case 32:
			goto st294
		case 44:
			goto tr8
		case 61:
			goto tr12
		case 92:
			goto tr10
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st294
		}
		goto tr6
tr496:
	( m.cs) = 296
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
tr491:
	( m.cs) = 296
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st296:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof296
		}
	st_case_296:
//line plugins/parsers/influx/machine.go:6762
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto tr496
		case 13:
			goto st34
		case 32:
			goto tr495
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto tr45
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr495
		}
		goto tr41
	st297:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof297
		}
	st_case_297:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr491
		case 13:
			goto tr453
		case 32:
			goto tr490
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st298
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr490
		}
		goto st12
	st298:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof298
		}
	st_case_298:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr491
		case 13:
			goto tr453
		case 32:
			goto tr490
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st299
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr490
		}
		goto st12
	st299:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof299
		}
	st_case_299:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr491
		case 13:
			goto tr453
		case 32:
			goto tr490
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st300
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr490
		}
		goto st12
	st300:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof300
		}
	st_case_300:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr491
		case 13:
			goto tr453
		case 32:
			goto tr490
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st301
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr490
		}
		goto st12
	st301:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof301
		}
	st_case_301:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr491
		case 13:
			goto tr453
		case 32:
			goto tr490
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st302
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr490
		}
		goto st12
	st302:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof302
		}
	st_case_302:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr491
		case 13:
			goto tr453
		case 32:
			goto tr490
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st303
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr490
		}
		goto st12
	st303:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof303
		}
	st_case_303:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr491
		case 13:
			goto tr453
		case 32:
			goto tr490
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st304
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr490
		}
		goto st12
	st304:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof304
		}
	st_case_304:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr491
		case 13:
			goto tr453
		case 32:
			goto tr490
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st305
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr490
		}
		goto st12
	st305:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof305
		}
	st_case_305:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr491
		case 13:
			goto tr453
		case 32:
			goto tr490
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st306
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr490
		}
		goto st12
	st306:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof306
		}
	st_case_306:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr491
		case 13:
			goto tr453
		case 32:
			goto tr490
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st307
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr490
		}
		goto st12
	st307:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof307
		}
	st_case_307:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr491
		case 13:
			goto tr453
		case 32:
			goto tr490
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st308
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr490
		}
		goto st12
	st308:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof308
		}
	st_case_308:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr491
		case 13:
			goto tr453
		case 32:
			goto tr490
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st309
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr490
		}
		goto st12
	st309:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof309
		}
	st_case_309:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr491
		case 13:
			goto tr453
		case 32:
			goto tr490
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st310
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr490
		}
		goto st12
	st310:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof310
		}
	st_case_310:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr491
		case 13:
			goto tr453
		case 32:
			goto tr490
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st311
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr490
		}
		goto st12
	st311:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof311
		}
	st_case_311:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr491
		case 13:
			goto tr453
		case 32:
			goto tr490
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st312
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr490
		}
		goto st12
	st312:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof312
		}
	st_case_312:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr491
		case 13:
			goto tr453
		case 32:
			goto tr490
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st313
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr490
		}
		goto st12
	st313:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof313
		}
	st_case_313:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr491
		case 13:
			goto tr453
		case 32:
			goto tr490
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st314
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr490
		}
		goto st12
	st314:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof314
		}
	st_case_314:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr491
		case 13:
			goto tr453
		case 32:
			goto tr490
		case 44:
			goto tr4
		case 61:
			goto tr49
		case 92:
			goto st29
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr490
		}
		goto st12
tr484:
	( m.cs) = 45
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr549:
	( m.cs) = 45
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr799:
	( m.cs) = 45
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr718:
	( m.cs) = 45
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr928:
	( m.cs) = 45
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr933:
	( m.cs) = 45
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr938:
	( m.cs) = 45
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr982:
	( m.cs) = 45
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr985:
	( m.cs) = 45
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr988:
	( m.cs) = 45
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st45:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof45
		}
	st_case_45:
//line plugins/parsers/influx/machine.go:7533
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr47
		case 44:
			goto tr47
		case 61:
			goto tr47
		case 92:
			goto tr135
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr47
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr47
		}
		goto tr134
tr134:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st46
	st46:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof46
		}
	st_case_46:
//line plugins/parsers/influx/machine.go:7564
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr47
		case 44:
			goto tr47
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr47
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr47
		}
		goto st46
tr137:
//line plugins/parsers/influx/machine.go.rl:86

	key = m.text()

//line plugins/parsers/influx/machine.go.rl:99

	key = m.text()

	goto st47
	st47:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof47
		}
	st_case_47:
//line plugins/parsers/influx/machine.go:7599
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr47
		case 34:
			goto tr139
		case 44:
			goto tr47
		case 45:
			goto tr140
		case 46:
			goto tr141
		case 48:
			goto tr142
		case 61:
			goto tr47
		case 70:
			goto tr144
		case 84:
			goto tr145
		case 92:
			goto tr58
		case 102:
			goto tr146
		case 116:
			goto tr147
		}
		switch {
		case ( m.data)[( m.p)] < 12:
			if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 10 {
				goto tr47
			}
		case ( m.data)[( m.p)] > 13:
			if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr143
			}
		default:
			goto tr47
		}
		goto tr57
tr139:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st48
	st48:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof48
		}
	st_case_48:
//line plugins/parsers/influx/machine.go:7650
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr149
		case 10:
			goto tr24
		case 11:
			goto tr150
		case 12:
			goto tr60
		case 13:
			goto tr25
		case 32:
			goto tr149
		case 34:
			goto tr151
		case 44:
			goto tr152
		case 61:
			goto tr23
		case 92:
			goto tr153
		}
		goto tr148
tr148:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st49
	st49:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof49
		}
	st_case_49:
//line plugins/parsers/influx/machine.go:7685
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr155
		case 10:
			goto st7
		case 11:
			goto tr156
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr155
		case 34:
			goto tr157
		case 44:
			goto tr158
		case 61:
			goto st6
		case 92:
			goto st64
		}
		goto st49
tr180:
	( m.cs) = 50
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr155:
	( m.cs) = 50
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr149:
	( m.cs) = 50
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
	st50:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof50
		}
	st_case_50:
//line plugins/parsers/influx/machine.go:7757
		switch ( m.data)[( m.p)] {
		case 9:
			goto st50
		case 10:
			goto st7
		case 11:
			goto tr162
		case 12:
			goto st2
		case 13:
			goto st8
		case 32:
			goto st50
		case 34:
			goto tr97
		case 44:
			goto st6
		case 61:
			goto st6
		case 92:
			goto tr163
		}
		goto tr160
tr160:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st51
	st51:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof51
		}
	st_case_51:
//line plugins/parsers/influx/machine.go:7792
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		goto st51
tr165:
//line plugins/parsers/influx/machine.go.rl:99

	key = m.text()

	goto st52
	st52:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof52
		}
	st_case_52:
//line plugins/parsers/influx/machine.go:7825
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr107
		case 45:
			goto tr167
		case 46:
			goto tr168
		case 48:
			goto tr169
		case 70:
			goto tr171
		case 84:
			goto tr172
		case 92:
			goto st76
		case 102:
			goto tr173
		case 116:
			goto tr174
		}
		if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr170
		}
		goto st6
tr167:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st53
	st53:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof53
		}
	st_case_53:
//line plugins/parsers/influx/machine.go:7867
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 46:
			goto st54
		case 48:
			goto st621
		case 92:
			goto st76
		}
		if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st622
		}
		goto st6
tr168:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st54
	st54:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof54
		}
	st_case_54:
//line plugins/parsers/influx/machine.go:7899
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st315
		}
		goto st6
	st315:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof315
		}
	st_case_315:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr515
		case 12:
			goto tr516
		case 13:
			goto tr517
		case 32:
			goto tr514
		case 34:
			goto tr31
		case 44:
			goto tr518
		case 69:
			goto st175
		case 92:
			goto st76
		case 101:
			goto st175
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st315
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr514
		}
		goto st6
tr902:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st316
tr514:
	( m.cs) = 316
//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr908:
	( m.cs) = 316
//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr911:
	( m.cs) = 316
//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr915:
	( m.cs) = 316
//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st316:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof316
		}
	st_case_316:
//line plugins/parsers/influx/machine.go:8013
		switch ( m.data)[( m.p)] {
		case 10:
			goto st317
		case 12:
			goto st261
		case 13:
			goto st104
		case 32:
			goto st316
		case 34:
			goto tr31
		case 45:
			goto tr522
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr523
			}
		case ( m.data)[( m.p)] >= 9:
			goto st316
		}
		goto st6
tr650:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st317
tr659:
	( m.cs) = 317
//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr515:
	( m.cs) = 317
//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr722:
	( m.cs) = 317
//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr729:
	( m.cs) = 317
//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr736:
	( m.cs) = 317
//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st317:
//line plugins/parsers/influx/machine.go.rl:157

	m.lineno++
	m.sol = m.p
	m.sol++ // next char will be the first column in the line

//line plugins/parsers/influx/machine.go.rl:163

	( m.cs) = 715;
	{( m.p)++; goto _out }

		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof317
		}
	st_case_317:
//line plugins/parsers/influx/machine.go:8126
		switch ( m.data)[( m.p)] {
		case 9:
			goto st166
		case 10:
			goto st7
		case 11:
			goto tr339
		case 12:
			goto st9
		case 13:
			goto st8
		case 32:
			goto st166
		case 34:
			goto tr118
		case 35:
			goto st6
		case 44:
			goto st6
		case 92:
			goto tr340
		}
		goto tr337
tr337:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st55
	st55:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof55
		}
	st_case_55:
//line plugins/parsers/influx/machine.go:8161
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr180
		case 10:
			goto st7
		case 11:
			goto tr181
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr180
		case 34:
			goto tr91
		case 44:
			goto tr182
		case 92:
			goto st157
		}
		goto st55
tr181:
	( m.cs) = 56
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st56:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof56
		}
	st_case_56:
//line plugins/parsers/influx/machine.go:8201
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr180
		case 10:
			goto st7
		case 11:
			goto tr185
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr180
		case 34:
			goto tr124
		case 44:
			goto tr182
		case 61:
			goto st55
		case 92:
			goto tr186
		}
		goto tr184
tr184:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st57
	st57:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof57
		}
	st_case_57:
//line plugins/parsers/influx/machine.go:8236
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr180
		case 10:
			goto st7
		case 11:
			goto tr188
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr180
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		goto st57
tr188:
	( m.cs) = 58
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr185:
	( m.cs) = 58
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
	st58:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof58
		}
	st_case_58:
//line plugins/parsers/influx/machine.go:8295
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr180
		case 10:
			goto st7
		case 11:
			goto tr185
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr180
		case 34:
			goto tr124
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto tr186
		}
		goto tr184
tr182:
	( m.cs) = 59
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr158:
	( m.cs) = 59
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr152:
	( m.cs) = 59
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
	st59:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof59
		}
	st_case_59:
//line plugins/parsers/influx/machine.go:8367
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr47
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr192
		case 44:
			goto st6
		case 61:
			goto st6
		case 92:
			goto tr193
		}
		goto tr191
tr191:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st60
	st60:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof60
		}
	st_case_60:
//line plugins/parsers/influx/machine.go:8400
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr47
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr195
		case 44:
			goto st6
		case 61:
			goto tr196
		case 92:
			goto st71
		}
		goto st60
tr192:
	( m.cs) = 318
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr195:
	( m.cs) = 318
//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st318:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof318
		}
	st_case_318:
//line plugins/parsers/influx/machine.go:8457
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto st319
		case 13:
			goto st34
		case 32:
			goto st261
		case 44:
			goto st37
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st261
		}
		goto st15
	st319:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof319
		}
	st_case_319:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto st319
		case 13:
			goto st34
		case 32:
			goto st261
		case 44:
			goto tr198
		case 45:
			goto tr525
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr526
			}
		case ( m.data)[( m.p)] >= 9:
			goto st261
		}
		goto st15
tr525:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st61
	st61:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof61
		}
	st_case_61:
//line plugins/parsers/influx/machine.go:8521
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr198
		case 44:
			goto tr198
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] < 12:
			if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 10 {
				goto tr198
			}
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st320
			}
		default:
			goto tr198
		}
		goto st15
tr526:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st320
	st320:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof320
		}
	st_case_320:
//line plugins/parsers/influx/machine.go:8556
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr527
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr198
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st322
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st15
tr527:
	( m.cs) = 321
//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st321:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof321
		}
	st_case_321:
//line plugins/parsers/influx/machine.go:8600
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto st321
		case 13:
			goto st34
		case 32:
			goto st266
		case 44:
			goto tr2
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st266
		}
		goto st15
	st322:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof322
		}
	st_case_322:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr527
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr198
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st323
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st15
	st323:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof323
		}
	st_case_323:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr527
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr198
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st324
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st15
	st324:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof324
		}
	st_case_324:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr527
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr198
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st325
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st15
	st325:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof325
		}
	st_case_325:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr527
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr198
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st326
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st15
	st326:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof326
		}
	st_case_326:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr527
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr198
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st327
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st15
	st327:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof327
		}
	st_case_327:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr527
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr198
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st328
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st15
	st328:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof328
		}
	st_case_328:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr527
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr198
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st329
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st15
	st329:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof329
		}
	st_case_329:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr527
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr198
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st330
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st15
	st330:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof330
		}
	st_case_330:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr527
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr198
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st331
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st15
	st331:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof331
		}
	st_case_331:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr527
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr198
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st332
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st15
	st332:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof332
		}
	st_case_332:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr527
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr198
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st333
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st15
	st333:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof333
		}
	st_case_333:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr527
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr198
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st334
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st15
	st334:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof334
		}
	st_case_334:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr527
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr198
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st335
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st15
	st335:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof335
		}
	st_case_335:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr527
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr198
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st336
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st15
	st336:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof336
		}
	st_case_336:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr527
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr198
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st337
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st15
	st337:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof337
		}
	st_case_337:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr527
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr198
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st338
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st15
	st338:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof338
		}
	st_case_338:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr527
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr198
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st339
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st15
	st339:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof339
		}
	st_case_339:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr527
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr198
		case 61:
			goto tr55
		case 92:
			goto st25
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr450
		}
		goto st15
tr196:
//line plugins/parsers/influx/machine.go.rl:86

	key = m.text()

	goto st62
	st62:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof62
		}
	st_case_62:
//line plugins/parsers/influx/machine.go:9167
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr47
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr151
		case 44:
			goto st6
		case 61:
			goto st6
		case 92:
			goto tr153
		}
		goto tr148
tr151:
	( m.cs) = 340
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr157:
	( m.cs) = 340
//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st340:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof340
		}
	st_case_340:
//line plugins/parsers/influx/machine.go:9224
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto tr548
		case 13:
			goto st34
		case 32:
			goto tr547
		case 44:
			goto tr549
		case 61:
			goto tr132
		case 92:
			goto st23
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr547
		}
		goto st17
tr548:
	( m.cs) = 341
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr716:
	( m.cs) = 341
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr926:
	( m.cs) = 341
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr931:
	( m.cs) = 341
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr936:
	( m.cs) = 341
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st341:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof341
		}
	st_case_341:
//line plugins/parsers/influx/machine.go:9355
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto tr550
		case 13:
			goto st34
		case 32:
			goto tr547
		case 44:
			goto tr62
		case 45:
			goto tr551
		case 61:
			goto tr132
		case 92:
			goto tr66
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr552
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr547
		}
		goto tr64
tr575:
	( m.cs) = 342
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr550:
	( m.cs) = 342
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
	st342:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof342
		}
	st_case_342:
//line plugins/parsers/influx/machine.go:9418
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto tr550
		case 13:
			goto st34
		case 32:
			goto tr547
		case 44:
			goto tr62
		case 45:
			goto tr551
		case 61:
			goto tr12
		case 92:
			goto tr66
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr552
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr547
		}
		goto tr64
tr551:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st63
	st63:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof63
		}
	st_case_63:
//line plugins/parsers/influx/machine.go:9457
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr132
		case 11:
			goto tr68
		case 13:
			goto tr132
		case 32:
			goto tr60
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st343
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr60
		}
		goto st19
tr552:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st343
	st343:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof343
		}
	st_case_343:
//line plugins/parsers/influx/machine.go:9494
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr554
		case 13:
			goto tr453
		case 32:
			goto tr553
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st345
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr553
		}
		goto st19
tr557:
	( m.cs) = 344
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
tr554:
	( m.cs) = 344
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st344:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof344
		}
	st_case_344:
//line plugins/parsers/influx/machine.go:9565
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto tr557
		case 13:
			goto st34
		case 32:
			goto tr556
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto tr66
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr556
		}
		goto tr64
	st345:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof345
		}
	st_case_345:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr554
		case 13:
			goto tr453
		case 32:
			goto tr553
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st346
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr553
		}
		goto st19
	st346:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof346
		}
	st_case_346:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr554
		case 13:
			goto tr453
		case 32:
			goto tr553
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st347
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr553
		}
		goto st19
	st347:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof347
		}
	st_case_347:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr554
		case 13:
			goto tr453
		case 32:
			goto tr553
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st348
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr553
		}
		goto st19
	st348:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof348
		}
	st_case_348:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr554
		case 13:
			goto tr453
		case 32:
			goto tr553
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st349
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr553
		}
		goto st19
	st349:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof349
		}
	st_case_349:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr554
		case 13:
			goto tr453
		case 32:
			goto tr553
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st350
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr553
		}
		goto st19
	st350:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof350
		}
	st_case_350:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr554
		case 13:
			goto tr453
		case 32:
			goto tr553
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st351
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr553
		}
		goto st19
	st351:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof351
		}
	st_case_351:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr554
		case 13:
			goto tr453
		case 32:
			goto tr553
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st352
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr553
		}
		goto st19
	st352:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof352
		}
	st_case_352:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr554
		case 13:
			goto tr453
		case 32:
			goto tr553
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st353
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr553
		}
		goto st19
	st353:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof353
		}
	st_case_353:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr554
		case 13:
			goto tr453
		case 32:
			goto tr553
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st354
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr553
		}
		goto st19
	st354:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof354
		}
	st_case_354:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr554
		case 13:
			goto tr453
		case 32:
			goto tr553
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st355
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr553
		}
		goto st19
	st355:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof355
		}
	st_case_355:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr554
		case 13:
			goto tr453
		case 32:
			goto tr553
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st356
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr553
		}
		goto st19
	st356:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof356
		}
	st_case_356:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr554
		case 13:
			goto tr453
		case 32:
			goto tr553
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st357
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr553
		}
		goto st19
	st357:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof357
		}
	st_case_357:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr554
		case 13:
			goto tr453
		case 32:
			goto tr553
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st358
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr553
		}
		goto st19
	st358:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof358
		}
	st_case_358:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr554
		case 13:
			goto tr453
		case 32:
			goto tr553
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st359
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr553
		}
		goto st19
	st359:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof359
		}
	st_case_359:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr554
		case 13:
			goto tr453
		case 32:
			goto tr553
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st360
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr553
		}
		goto st19
	st360:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof360
		}
	st_case_360:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr554
		case 13:
			goto tr453
		case 32:
			goto tr553
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st361
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr553
		}
		goto st19
	st361:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof361
		}
	st_case_361:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr554
		case 13:
			goto tr453
		case 32:
			goto tr553
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st362
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr553
		}
		goto st19
	st362:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof362
		}
	st_case_362:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr554
		case 13:
			goto tr453
		case 32:
			goto tr553
		case 44:
			goto tr62
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr553
		}
		goto st19
tr153:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st64
	st64:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof64
		}
	st_case_64:
//line plugins/parsers/influx/machine.go:10132
		switch ( m.data)[( m.p)] {
		case 34:
			goto st49
		case 92:
			goto st65
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr47
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr47
		}
		goto st17
	st65:
//line plugins/parsers/influx/machine.go.rl:234
 ( m.p)--
 
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof65
		}
	st_case_65:
//line plugins/parsers/influx/machine.go:10156
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr155
		case 10:
			goto st7
		case 11:
			goto tr156
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr155
		case 34:
			goto tr157
		case 44:
			goto tr158
		case 61:
			goto st6
		case 92:
			goto st64
		}
		goto st49
tr156:
	( m.cs) = 66
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr150:
	( m.cs) = 66
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
	st66:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof66
		}
	st_case_66:
//line plugins/parsers/influx/machine.go:10215
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr155
		case 10:
			goto st7
		case 11:
			goto tr203
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr155
		case 34:
			goto tr204
		case 44:
			goto tr158
		case 61:
			goto st6
		case 92:
			goto tr205
		}
		goto tr202
tr202:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st67
	st67:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof67
		}
	st_case_67:
//line plugins/parsers/influx/machine.go:10250
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr155
		case 10:
			goto st7
		case 11:
			goto tr207
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr155
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		goto st67
tr207:
	( m.cs) = 68
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr203:
	( m.cs) = 68
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
	st68:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof68
		}
	st_case_68:
//line plugins/parsers/influx/machine.go:10309
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr155
		case 10:
			goto st7
		case 11:
			goto tr203
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr155
		case 34:
			goto tr204
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto tr205
		}
		goto tr202
tr204:
	( m.cs) = 363
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr208:
	( m.cs) = 363
//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st363:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof363
		}
	st_case_363:
//line plugins/parsers/influx/machine.go:10368
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto tr575
		case 13:
			goto st34
		case 32:
			goto tr547
		case 44:
			goto tr549
		case 61:
			goto tr12
		case 92:
			goto st21
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr547
		}
		goto st19
tr205:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st69
	st69:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof69
		}
	st_case_69:
//line plugins/parsers/influx/machine.go:10400
		switch ( m.data)[( m.p)] {
		case 34:
			goto st67
		case 92:
			goto st70
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr47
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr47
		}
		goto st19
	st70:
//line plugins/parsers/influx/machine.go.rl:234
 ( m.p)--
 
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof70
		}
	st_case_70:
//line plugins/parsers/influx/machine.go:10424
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr155
		case 10:
			goto st7
		case 11:
			goto tr207
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr155
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		goto st67
tr193:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st71
	st71:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof71
		}
	st_case_71:
//line plugins/parsers/influx/machine.go:10459
		switch ( m.data)[( m.p)] {
		case 34:
			goto st60
		case 92:
			goto st72
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr47
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr47
		}
		goto st15
	st72:
//line plugins/parsers/influx/machine.go.rl:234
 ( m.p)--
 
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof72
		}
	st_case_72:
//line plugins/parsers/influx/machine.go:10483
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr47
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr195
		case 44:
			goto st6
		case 61:
			goto tr196
		case 92:
			goto st71
		}
		goto st60
tr189:
//line plugins/parsers/influx/machine.go.rl:99

	key = m.text()

	goto st73
tr346:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:99

	key = m.text()

	goto st73
	st73:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof73
		}
	st_case_73:
//line plugins/parsers/influx/machine.go:10526
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr180
		case 10:
			goto st7
		case 11:
			goto tr181
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr180
		case 34:
			goto tr212
		case 44:
			goto tr182
		case 45:
			goto tr213
		case 46:
			goto tr214
		case 48:
			goto tr215
		case 70:
			goto tr217
		case 84:
			goto tr218
		case 92:
			goto st157
		case 102:
			goto tr219
		case 116:
			goto tr220
		}
		if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr216
		}
		goto st55
tr212:
	( m.cs) = 364
//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st364:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof364
		}
	st_case_364:
//line plugins/parsers/influx/machine.go:10583
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr576
		case 10:
			goto tr475
		case 11:
			goto tr577
		case 12:
			goto tr482
		case 13:
			goto tr476
		case 32:
			goto tr576
		case 34:
			goto tr85
		case 44:
			goto tr578
		case 92:
			goto tr87
		}
		goto tr82
tr607:
	( m.cs) = 365
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr576:
	( m.cs) = 365
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
tr749:
	( m.cs) = 365
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr619:
	( m.cs) = 365
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr745:
	( m.cs) = 365
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr777:
	( m.cs) = 365
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr783:
	( m.cs) = 365
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr789:
	( m.cs) = 365
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr802:
	( m.cs) = 365
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr807:
	( m.cs) = 365
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr812:
	( m.cs) = 365
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st365:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof365
		}
	st_case_365:
//line plugins/parsers/influx/machine.go:10837
		switch ( m.data)[( m.p)] {
		case 9:
			goto st365
		case 10:
			goto st288
		case 11:
			goto tr580
		case 12:
			goto st290
		case 13:
			goto st74
		case 32:
			goto st365
		case 34:
			goto tr97
		case 44:
			goto st6
		case 45:
			goto tr581
		case 61:
			goto st6
		case 92:
			goto tr98
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr582
		}
		goto tr94
tr580:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st366
	st366:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof366
		}
	st_case_366:
//line plugins/parsers/influx/machine.go:10877
		switch ( m.data)[( m.p)] {
		case 9:
			goto st365
		case 10:
			goto st288
		case 11:
			goto tr580
		case 12:
			goto st290
		case 13:
			goto st74
		case 32:
			goto st365
		case 34:
			goto tr97
		case 44:
			goto st6
		case 45:
			goto tr581
		case 61:
			goto tr101
		case 92:
			goto tr98
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr582
		}
		goto tr94
tr476:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st74
tr586:
	( m.cs) = 74
//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr623:
	( m.cs) = 74
//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr780:
	( m.cs) = 74
//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr786:
	( m.cs) = 74
//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr792:
	( m.cs) = 74
//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st74:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof74
		}
	st_case_74:
//line plugins/parsers/influx/machine.go:10982
		if ( m.data)[( m.p)] == 10 {
			goto st288
		}
		goto tr8
tr581:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st75
	st75:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof75
		}
	st_case_75:
//line plugins/parsers/influx/machine.go:10998
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr105
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st367
		}
		goto st33
tr582:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st367
	st367:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof367
		}
	st_case_367:
//line plugins/parsers/influx/machine.go:11034
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr583
		case 10:
			goto tr584
		case 11:
			goto tr585
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st370
		}
		goto st33
tr583:
	( m.cs) = 368
//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st368:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof368
		}
	st_case_368:
//line plugins/parsers/influx/machine.go:11079
		switch ( m.data)[( m.p)] {
		case 10:
			goto st288
		case 12:
			goto st266
		case 13:
			goto st74
		case 32:
			goto st368
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto st368
		}
		goto st6
tr27:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st76
	st76:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof76
		}
	st_case_76:
//line plugins/parsers/influx/machine.go:11109
		switch ( m.data)[( m.p)] {
		case 34:
			goto st6
		case 92:
			goto st6
		}
		goto tr8
tr585:
	( m.cs) = 369
//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st369:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof369
		}
	st_case_369:
//line plugins/parsers/influx/machine.go:11135
		switch ( m.data)[( m.p)] {
		case 9:
			goto st368
		case 10:
			goto st288
		case 11:
			goto st369
		case 12:
			goto st266
		case 13:
			goto st74
		case 32:
			goto st368
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		goto st33
tr98:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st77
	st77:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof77
		}
	st_case_77:
//line plugins/parsers/influx/machine.go:11170
		switch ( m.data)[( m.p)] {
		case 34:
			goto st33
		case 92:
			goto st33
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr8
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr8
		}
		goto st3
	st370:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof370
		}
	st_case_370:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr583
		case 10:
			goto tr584
		case 11:
			goto tr585
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st371
		}
		goto st33
	st371:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof371
		}
	st_case_371:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr583
		case 10:
			goto tr584
		case 11:
			goto tr585
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st372
		}
		goto st33
	st372:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof372
		}
	st_case_372:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr583
		case 10:
			goto tr584
		case 11:
			goto tr585
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st373
		}
		goto st33
	st373:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof373
		}
	st_case_373:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr583
		case 10:
			goto tr584
		case 11:
			goto tr585
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st374
		}
		goto st33
	st374:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof374
		}
	st_case_374:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr583
		case 10:
			goto tr584
		case 11:
			goto tr585
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st375
		}
		goto st33
	st375:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof375
		}
	st_case_375:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr583
		case 10:
			goto tr584
		case 11:
			goto tr585
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st376
		}
		goto st33
	st376:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof376
		}
	st_case_376:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr583
		case 10:
			goto tr584
		case 11:
			goto tr585
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st377
		}
		goto st33
	st377:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof377
		}
	st_case_377:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr583
		case 10:
			goto tr584
		case 11:
			goto tr585
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st378
		}
		goto st33
	st378:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof378
		}
	st_case_378:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr583
		case 10:
			goto tr584
		case 11:
			goto tr585
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st379
		}
		goto st33
	st379:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof379
		}
	st_case_379:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr583
		case 10:
			goto tr584
		case 11:
			goto tr585
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st380
		}
		goto st33
	st380:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof380
		}
	st_case_380:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr583
		case 10:
			goto tr584
		case 11:
			goto tr585
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st381
		}
		goto st33
	st381:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof381
		}
	st_case_381:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr583
		case 10:
			goto tr584
		case 11:
			goto tr585
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st382
		}
		goto st33
	st382:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof382
		}
	st_case_382:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr583
		case 10:
			goto tr584
		case 11:
			goto tr585
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st383
		}
		goto st33
	st383:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof383
		}
	st_case_383:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr583
		case 10:
			goto tr584
		case 11:
			goto tr585
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st384
		}
		goto st33
	st384:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof384
		}
	st_case_384:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr583
		case 10:
			goto tr584
		case 11:
			goto tr585
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st385
		}
		goto st33
	st385:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof385
		}
	st_case_385:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr583
		case 10:
			goto tr584
		case 11:
			goto tr585
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st386
		}
		goto st33
	st386:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof386
		}
	st_case_386:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr583
		case 10:
			goto tr584
		case 11:
			goto tr585
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st387
		}
		goto st33
	st387:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof387
		}
	st_case_387:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr583
		case 10:
			goto tr584
		case 11:
			goto tr585
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto st77
		}
		goto st33
tr577:
	( m.cs) = 388
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
tr621:
	( m.cs) = 388
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr803:
	( m.cs) = 388
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr808:
	( m.cs) = 388
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr813:
	( m.cs) = 388
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st388:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof388
		}
	st_case_388:
//line plugins/parsers/influx/machine.go:11855
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr607
		case 10:
			goto st288
		case 11:
			goto tr608
		case 12:
			goto tr482
		case 13:
			goto st74
		case 32:
			goto tr607
		case 34:
			goto tr124
		case 44:
			goto tr92
		case 45:
			goto tr609
		case 61:
			goto st31
		case 92:
			goto tr125
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr610
		}
		goto tr121
tr608:
	( m.cs) = 389
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
	st389:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof389
		}
	st_case_389:
//line plugins/parsers/influx/machine.go:11906
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr607
		case 10:
			goto st288
		case 11:
			goto tr608
		case 12:
			goto tr482
		case 13:
			goto st74
		case 32:
			goto tr607
		case 34:
			goto tr124
		case 44:
			goto tr92
		case 45:
			goto tr609
		case 61:
			goto tr129
		case 92:
			goto tr125
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr610
		}
		goto tr121
tr92:
	( m.cs) = 78
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr86:
	( m.cs) = 78
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
tr233:
	( m.cs) = 78
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st78:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof78
		}
	st_case_78:
//line plugins/parsers/influx/machine.go:11983
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr47
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr192
		case 44:
			goto st6
		case 61:
			goto st6
		case 92:
			goto tr224
		}
		goto tr223
tr223:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st79
	st79:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof79
		}
	st_case_79:
//line plugins/parsers/influx/machine.go:12016
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr47
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr195
		case 44:
			goto st6
		case 61:
			goto tr226
		case 92:
			goto st89
		}
		goto st79
tr226:
//line plugins/parsers/influx/machine.go.rl:86

	key = m.text()

	goto st80
	st80:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof80
		}
	st_case_80:
//line plugins/parsers/influx/machine.go:12049
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr47
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr151
		case 44:
			goto st6
		case 61:
			goto st6
		case 92:
			goto tr229
		}
		goto tr228
tr228:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st81
	st81:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof81
		}
	st_case_81:
//line plugins/parsers/influx/machine.go:12082
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr231
		case 10:
			goto st7
		case 11:
			goto tr232
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr231
		case 34:
			goto tr157
		case 44:
			goto tr233
		case 61:
			goto st6
		case 92:
			goto st87
		}
		goto st81
tr232:
	( m.cs) = 82
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st82:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof82
		}
	st_case_82:
//line plugins/parsers/influx/machine.go:12124
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr231
		case 10:
			goto st7
		case 11:
			goto tr236
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr231
		case 34:
			goto tr204
		case 44:
			goto tr233
		case 61:
			goto st6
		case 92:
			goto tr237
		}
		goto tr235
tr235:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st83
	st83:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof83
		}
	st_case_83:
//line plugins/parsers/influx/machine.go:12159
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr231
		case 10:
			goto st7
		case 11:
			goto tr239
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr231
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		goto st83
tr239:
	( m.cs) = 84
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr236:
	( m.cs) = 84
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
	st84:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof84
		}
	st_case_84:
//line plugins/parsers/influx/machine.go:12218
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr231
		case 10:
			goto st7
		case 11:
			goto tr236
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr231
		case 34:
			goto tr204
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto tr237
		}
		goto tr235
tr237:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st85
	st85:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof85
		}
	st_case_85:
//line plugins/parsers/influx/machine.go:12253
		switch ( m.data)[( m.p)] {
		case 34:
			goto st83
		case 92:
			goto st86
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr47
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr47
		}
		goto st19
	st86:
//line plugins/parsers/influx/machine.go.rl:234
 ( m.p)--
 
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof86
		}
	st_case_86:
//line plugins/parsers/influx/machine.go:12277
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr231
		case 10:
			goto st7
		case 11:
			goto tr239
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr231
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		goto st83
tr229:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st87
	st87:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof87
		}
	st_case_87:
//line plugins/parsers/influx/machine.go:12312
		switch ( m.data)[( m.p)] {
		case 34:
			goto st81
		case 92:
			goto st88
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr47
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr47
		}
		goto st17
	st88:
//line plugins/parsers/influx/machine.go.rl:234
 ( m.p)--
 
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof88
		}
	st_case_88:
//line plugins/parsers/influx/machine.go:12336
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr231
		case 10:
			goto st7
		case 11:
			goto tr232
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr231
		case 34:
			goto tr157
		case 44:
			goto tr233
		case 61:
			goto st6
		case 92:
			goto st87
		}
		goto st81
tr224:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st89
	st89:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof89
		}
	st_case_89:
//line plugins/parsers/influx/machine.go:12371
		switch ( m.data)[( m.p)] {
		case 34:
			goto st79
		case 92:
			goto st90
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr47
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr47
		}
		goto st15
	st90:
//line plugins/parsers/influx/machine.go.rl:234
 ( m.p)--
 
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof90
		}
	st_case_90:
//line plugins/parsers/influx/machine.go:12395
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr47
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr195
		case 44:
			goto st6
		case 61:
			goto tr226
		case 92:
			goto st89
		}
		goto st79
tr609:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st91
	st91:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof91
		}
	st_case_91:
//line plugins/parsers/influx/machine.go:12428
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr89
		case 10:
			goto st7
		case 11:
			goto tr127
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr89
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st390
		}
		goto st42
tr610:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st390
	st390:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof390
		}
	st_case_390:
//line plugins/parsers/influx/machine.go:12466
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr611
		case 10:
			goto tr584
		case 11:
			goto tr612
		case 12:
			goto tr490
		case 13:
			goto tr586
		case 32:
			goto tr611
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st534
		}
		goto st42
tr616:
	( m.cs) = 391
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr756:
	( m.cs) = 391
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr611:
	( m.cs) = 391
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr753:
	( m.cs) = 391
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st391:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof391
		}
	st_case_391:
//line plugins/parsers/influx/machine.go:12570
		switch ( m.data)[( m.p)] {
		case 9:
			goto st391
		case 10:
			goto st288
		case 11:
			goto tr615
		case 12:
			goto st294
		case 13:
			goto st74
		case 32:
			goto st391
		case 34:
			goto tr97
		case 44:
			goto st6
		case 61:
			goto st6
		case 92:
			goto tr98
		}
		goto tr94
tr615:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st392
	st392:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof392
		}
	st_case_392:
//line plugins/parsers/influx/machine.go:12605
		switch ( m.data)[( m.p)] {
		case 9:
			goto st391
		case 10:
			goto st288
		case 11:
			goto tr615
		case 12:
			goto st294
		case 13:
			goto st74
		case 32:
			goto st391
		case 34:
			goto tr97
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto tr98
		}
		goto tr94
tr617:
	( m.cs) = 393
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
tr612:
	( m.cs) = 393
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st393:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof393
		}
	st_case_393:
//line plugins/parsers/influx/machine.go:12674
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr616
		case 10:
			goto st288
		case 11:
			goto tr617
		case 12:
			goto tr495
		case 13:
			goto st74
		case 32:
			goto tr616
		case 34:
			goto tr124
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto tr125
		}
		goto tr121
tr129:
//line plugins/parsers/influx/machine.go.rl:99

	key = m.text()

	goto st92
tr374:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:99

	key = m.text()

	goto st92
	st92:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof92
		}
	st_case_92:
//line plugins/parsers/influx/machine.go:12719
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr89
		case 10:
			goto st7
		case 11:
			goto tr90
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr89
		case 34:
			goto tr212
		case 44:
			goto tr92
		case 45:
			goto tr245
		case 46:
			goto tr246
		case 48:
			goto tr247
		case 70:
			goto tr249
		case 84:
			goto tr250
		case 92:
			goto st142
		case 102:
			goto tr251
		case 116:
			goto tr252
		}
		if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr248
		}
		goto st31
tr90:
	( m.cs) = 93
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr84:
	( m.cs) = 93
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
	st93:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof93
		}
	st_case_93:
//line plugins/parsers/influx/machine.go:12793
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr89
		case 10:
			goto st7
		case 11:
			goto tr131
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr89
		case 34:
			goto tr124
		case 44:
			goto tr92
		case 61:
			goto st31
		case 92:
			goto tr125
		}
		goto tr121
tr125:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st94
	st94:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof94
		}
	st_case_94:
//line plugins/parsers/influx/machine.go:12828
		switch ( m.data)[( m.p)] {
		case 34:
			goto st42
		case 92:
			goto st42
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr8
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr8
		}
		goto st12
tr245:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st95
	st95:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof95
		}
	st_case_95:
//line plugins/parsers/influx/machine.go:12855
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr89
		case 10:
			goto st7
		case 11:
			goto tr90
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr89
		case 34:
			goto tr91
		case 44:
			goto tr92
		case 46:
			goto st97
		case 48:
			goto st522
		case 92:
			goto st142
		}
		if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st525
		}
		goto st31
tr85:
	( m.cs) = 394
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr91:
	( m.cs) = 394
//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr118:
	( m.cs) = 394
//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
	st394:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof394
		}
	st_case_394:
//line plugins/parsers/influx/machine.go:12936
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto tr618
		case 13:
			goto st34
		case 32:
			goto tr482
		case 44:
			goto tr484
		case 92:
			goto st96
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr482
		}
		goto st1
tr618:
	( m.cs) = 395
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr798:
	( m.cs) = 395
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr981:
	( m.cs) = 395
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr984:
	( m.cs) = 395
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr987:
	( m.cs) = 395
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st395:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof395
		}
	st_case_395:
//line plugins/parsers/influx/machine.go:13065
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto tr487
		case 13:
			goto st34
		case 32:
			goto tr482
		case 44:
			goto tr4
		case 45:
			goto tr488
		case 61:
			goto st1
		case 92:
			goto tr45
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr489
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr482
		}
		goto tr41
tr37:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st96
tr441:
//line plugins/parsers/influx/machine.go.rl:73

	foundMetric = true

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st96
	st96:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof96
		}
	st_case_96:
//line plugins/parsers/influx/machine.go:13114
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto st0
			}
		case ( m.data)[( m.p)] >= 9:
			goto st0
		}
		goto st1
tr246:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st97
	st97:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof97
		}
	st_case_97:
//line plugins/parsers/influx/machine.go:13135
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr89
		case 10:
			goto st7
		case 11:
			goto tr90
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr89
		case 34:
			goto tr91
		case 44:
			goto tr92
		case 92:
			goto st142
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st396
		}
		goto st31
	st396:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof396
		}
	st_case_396:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr619
		case 10:
			goto tr620
		case 11:
			goto tr621
		case 12:
			goto tr622
		case 13:
			goto tr623
		case 32:
			goto tr619
		case 34:
			goto tr91
		case 44:
			goto tr624
		case 69:
			goto st140
		case 92:
			goto st142
		case 101:
			goto st140
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st396
		}
		goto st31
tr578:
	( m.cs) = 98
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
tr624:
	( m.cs) = 98
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr747:
	( m.cs) = 98
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr781:
	( m.cs) = 98
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr787:
	( m.cs) = 98
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr793:
	( m.cs) = 98
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr805:
	( m.cs) = 98
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr810:
	( m.cs) = 98
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr815:
	( m.cs) = 98
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st98:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof98
		}
	st_case_98:
//line plugins/parsers/influx/machine.go:13399
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr47
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr258
		case 44:
			goto st6
		case 61:
			goto st6
		case 92:
			goto tr259
		}
		goto tr257
tr257:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st99
	st99:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof99
		}
	st_case_99:
//line plugins/parsers/influx/machine.go:13432
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr47
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr261
		case 44:
			goto st6
		case 61:
			goto tr262
		case 92:
			goto st138
		}
		goto st99
tr258:
	( m.cs) = 397
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr261:
	( m.cs) = 397
//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st397:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof397
		}
	st_case_397:
//line plugins/parsers/influx/machine.go:13489
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto st398
		case 13:
			goto st34
		case 32:
			goto st261
		case 44:
			goto st37
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st261
		}
		goto st46
	st398:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof398
		}
	st_case_398:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto st398
		case 13:
			goto st34
		case 32:
			goto st261
		case 44:
			goto tr132
		case 45:
			goto tr627
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto tr628
			}
		case ( m.data)[( m.p)] >= 9:
			goto st261
		}
		goto st46
tr627:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st100
	st100:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof100
		}
	st_case_100:
//line plugins/parsers/influx/machine.go:13553
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr132
		case 44:
			goto tr132
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] < 12:
			if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 10 {
				goto tr132
			}
		case ( m.data)[( m.p)] > 13:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st399
			}
		default:
			goto tr132
		}
		goto st46
tr628:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st399
	st399:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof399
		}
	st_case_399:
//line plugins/parsers/influx/machine.go:13588
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr629
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr132
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st401
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st46
tr629:
	( m.cs) = 400
//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st400:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof400
		}
	st_case_400:
//line plugins/parsers/influx/machine.go:13632
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto st400
		case 13:
			goto st34
		case 32:
			goto st266
		case 44:
			goto tr47
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st266
		}
		goto st46
tr135:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st101
	st101:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof101
		}
	st_case_101:
//line plugins/parsers/influx/machine.go:13664
		if ( m.data)[( m.p)] == 92 {
			goto st102
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr47
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr47
		}
		goto st46
	st102:
//line plugins/parsers/influx/machine.go.rl:234
 ( m.p)--
 
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof102
		}
	st_case_102:
//line plugins/parsers/influx/machine.go:13685
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr47
		case 44:
			goto tr47
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr47
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr47
		}
		goto st46
	st401:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof401
		}
	st_case_401:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr629
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr132
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st402
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st46
	st402:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof402
		}
	st_case_402:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr629
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr132
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st403
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st46
	st403:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof403
		}
	st_case_403:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr629
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr132
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st404
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st46
	st404:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof404
		}
	st_case_404:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr629
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr132
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st405
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st46
	st405:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof405
		}
	st_case_405:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr629
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr132
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st406
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st46
	st406:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof406
		}
	st_case_406:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr629
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr132
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st407
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st46
	st407:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof407
		}
	st_case_407:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr629
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr132
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st408
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st46
	st408:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof408
		}
	st_case_408:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr629
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr132
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st409
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st46
	st409:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof409
		}
	st_case_409:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr629
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr132
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st410
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st46
	st410:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof410
		}
	st_case_410:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr629
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr132
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st411
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st46
	st411:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof411
		}
	st_case_411:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr629
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr132
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st412
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st46
	st412:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof412
		}
	st_case_412:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr629
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr132
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st413
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st46
	st413:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof413
		}
	st_case_413:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr629
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr132
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st414
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st46
	st414:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof414
		}
	st_case_414:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr629
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr132
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st415
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st46
	st415:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof415
		}
	st_case_415:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr629
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr132
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st416
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st46
	st416:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof416
		}
	st_case_416:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr629
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr132
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st417
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st46
	st417:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof417
		}
	st_case_417:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr629
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr132
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st418
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto st46
	st418:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof418
		}
	st_case_418:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 11:
			goto tr629
		case 13:
			goto tr453
		case 32:
			goto tr450
		case 44:
			goto tr132
		case 61:
			goto tr137
		case 92:
			goto st101
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr450
		}
		goto st46
tr262:
//line plugins/parsers/influx/machine.go.rl:86

	key = m.text()

//line plugins/parsers/influx/machine.go.rl:99

	key = m.text()

	goto st103
	st103:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof103
		}
	st_case_103:
//line plugins/parsers/influx/machine.go:14255
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr47
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr266
		case 44:
			goto st6
		case 45:
			goto tr267
		case 46:
			goto tr268
		case 48:
			goto tr269
		case 61:
			goto st6
		case 70:
			goto tr271
		case 84:
			goto tr272
		case 92:
			goto tr229
		case 102:
			goto tr273
		case 116:
			goto tr274
		}
		if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr270
		}
		goto tr228
tr266:
	( m.cs) = 419
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st419:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof419
		}
	st_case_419:
//line plugins/parsers/influx/machine.go:14316
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr649
		case 10:
			goto tr650
		case 11:
			goto tr651
		case 12:
			goto tr547
		case 13:
			goto tr652
		case 32:
			goto tr649
		case 34:
			goto tr151
		case 44:
			goto tr653
		case 61:
			goto tr23
		case 92:
			goto tr153
		}
		goto tr148
tr841:
	( m.cs) = 420
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr682:
	( m.cs) = 420
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr649:
	( m.cs) = 420
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
tr837:
	( m.cs) = 420
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr710:
	( m.cs) = 420
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr721:
	( m.cs) = 420
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr728:
	( m.cs) = 420
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr735:
	( m.cs) = 420
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr869:
	( m.cs) = 420
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr873:
	( m.cs) = 420
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr877:
	( m.cs) = 420
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st420:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof420
		}
	st_case_420:
//line plugins/parsers/influx/machine.go:14572
		switch ( m.data)[( m.p)] {
		case 9:
			goto st420
		case 10:
			goto st317
		case 11:
			goto tr655
		case 12:
			goto st290
		case 13:
			goto st104
		case 32:
			goto st420
		case 34:
			goto tr97
		case 44:
			goto st6
		case 45:
			goto tr656
		case 61:
			goto st6
		case 92:
			goto tr163
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr657
		}
		goto tr160
tr655:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st421
	st421:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof421
		}
	st_case_421:
//line plugins/parsers/influx/machine.go:14612
		switch ( m.data)[( m.p)] {
		case 9:
			goto st420
		case 10:
			goto st317
		case 11:
			goto tr655
		case 12:
			goto st290
		case 13:
			goto st104
		case 32:
			goto st420
		case 34:
			goto tr97
		case 44:
			goto st6
		case 45:
			goto tr656
		case 61:
			goto tr165
		case 92:
			goto tr163
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr657
		}
		goto tr160
tr652:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st104
tr661:
	( m.cs) = 104
//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr517:
	( m.cs) = 104
//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr725:
	( m.cs) = 104
//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr732:
	( m.cs) = 104
//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr739:
	( m.cs) = 104
//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st104:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof104
		}
	st_case_104:
//line plugins/parsers/influx/machine.go:14717
		if ( m.data)[( m.p)] == 10 {
			goto st317
		}
		goto tr8
tr656:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st105
	st105:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof105
		}
	st_case_105:
//line plugins/parsers/influx/machine.go:14733
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr105
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st422
		}
		goto st51
tr657:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st422
	st422:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof422
		}
	st_case_422:
//line plugins/parsers/influx/machine.go:14769
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr658
		case 10:
			goto tr659
		case 11:
			goto tr660
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st425
		}
		goto st51
tr658:
	( m.cs) = 423
//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st423:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof423
		}
	st_case_423:
//line plugins/parsers/influx/machine.go:14814
		switch ( m.data)[( m.p)] {
		case 10:
			goto st317
		case 12:
			goto st266
		case 13:
			goto st104
		case 32:
			goto st423
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto st423
		}
		goto st6
tr660:
	( m.cs) = 424
//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st424:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof424
		}
	st_case_424:
//line plugins/parsers/influx/machine.go:14851
		switch ( m.data)[( m.p)] {
		case 9:
			goto st423
		case 10:
			goto st317
		case 11:
			goto st424
		case 12:
			goto st266
		case 13:
			goto st104
		case 32:
			goto st423
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		goto st51
tr163:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st106
	st106:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof106
		}
	st_case_106:
//line plugins/parsers/influx/machine.go:14886
		switch ( m.data)[( m.p)] {
		case 34:
			goto st51
		case 92:
			goto st51
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr8
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr8
		}
		goto st3
	st425:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof425
		}
	st_case_425:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr658
		case 10:
			goto tr659
		case 11:
			goto tr660
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st426
		}
		goto st51
	st426:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof426
		}
	st_case_426:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr658
		case 10:
			goto tr659
		case 11:
			goto tr660
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st427
		}
		goto st51
	st427:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof427
		}
	st_case_427:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr658
		case 10:
			goto tr659
		case 11:
			goto tr660
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st428
		}
		goto st51
	st428:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof428
		}
	st_case_428:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr658
		case 10:
			goto tr659
		case 11:
			goto tr660
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st429
		}
		goto st51
	st429:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof429
		}
	st_case_429:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr658
		case 10:
			goto tr659
		case 11:
			goto tr660
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st430
		}
		goto st51
	st430:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof430
		}
	st_case_430:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr658
		case 10:
			goto tr659
		case 11:
			goto tr660
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st431
		}
		goto st51
	st431:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof431
		}
	st_case_431:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr658
		case 10:
			goto tr659
		case 11:
			goto tr660
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st432
		}
		goto st51
	st432:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof432
		}
	st_case_432:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr658
		case 10:
			goto tr659
		case 11:
			goto tr660
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st433
		}
		goto st51
	st433:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof433
		}
	st_case_433:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr658
		case 10:
			goto tr659
		case 11:
			goto tr660
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st434
		}
		goto st51
	st434:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof434
		}
	st_case_434:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr658
		case 10:
			goto tr659
		case 11:
			goto tr660
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st435
		}
		goto st51
	st435:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof435
		}
	st_case_435:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr658
		case 10:
			goto tr659
		case 11:
			goto tr660
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st436
		}
		goto st51
	st436:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof436
		}
	st_case_436:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr658
		case 10:
			goto tr659
		case 11:
			goto tr660
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st437
		}
		goto st51
	st437:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof437
		}
	st_case_437:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr658
		case 10:
			goto tr659
		case 11:
			goto tr660
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st438
		}
		goto st51
	st438:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof438
		}
	st_case_438:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr658
		case 10:
			goto tr659
		case 11:
			goto tr660
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st439
		}
		goto st51
	st439:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof439
		}
	st_case_439:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr658
		case 10:
			goto tr659
		case 11:
			goto tr660
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st440
		}
		goto st51
	st440:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof440
		}
	st_case_440:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr658
		case 10:
			goto tr659
		case 11:
			goto tr660
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st441
		}
		goto st51
	st441:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof441
		}
	st_case_441:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr658
		case 10:
			goto tr659
		case 11:
			goto tr660
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st442
		}
		goto st51
	st442:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof442
		}
	st_case_442:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr658
		case 10:
			goto tr659
		case 11:
			goto tr660
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto st106
		}
		goto st51
tr651:
	( m.cs) = 443
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
tr711:
	( m.cs) = 443
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr723:
	( m.cs) = 443
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr730:
	( m.cs) = 443
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr737:
	( m.cs) = 443
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st443:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof443
		}
	st_case_443:
//line plugins/parsers/influx/machine.go:15571
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr682
		case 10:
			goto st317
		case 11:
			goto tr683
		case 12:
			goto tr547
		case 13:
			goto st104
		case 32:
			goto tr682
		case 34:
			goto tr204
		case 44:
			goto tr158
		case 45:
			goto tr684
		case 61:
			goto st6
		case 92:
			goto tr205
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr685
		}
		goto tr202
tr683:
	( m.cs) = 444
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
	st444:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof444
		}
	st_case_444:
//line plugins/parsers/influx/machine.go:15622
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr682
		case 10:
			goto st317
		case 11:
			goto tr683
		case 12:
			goto tr547
		case 13:
			goto st104
		case 32:
			goto tr682
		case 34:
			goto tr204
		case 44:
			goto tr158
		case 45:
			goto tr684
		case 61:
			goto tr165
		case 92:
			goto tr205
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr685
		}
		goto tr202
tr684:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st107
	st107:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof107
		}
	st_case_107:
//line plugins/parsers/influx/machine.go:15662
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr155
		case 10:
			goto st7
		case 11:
			goto tr207
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr155
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st445
		}
		goto st67
tr685:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st445
	st445:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof445
		}
	st_case_445:
//line plugins/parsers/influx/machine.go:15700
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr686
		case 10:
			goto tr659
		case 11:
			goto tr687
		case 12:
			goto tr553
		case 13:
			goto tr661
		case 32:
			goto tr686
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st449
		}
		goto st67
tr848:
	( m.cs) = 446
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr691:
	( m.cs) = 446
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr845:
	( m.cs) = 446
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr686:
	( m.cs) = 446
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st446:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof446
		}
	st_case_446:
//line plugins/parsers/influx/machine.go:15804
		switch ( m.data)[( m.p)] {
		case 9:
			goto st446
		case 10:
			goto st317
		case 11:
			goto tr690
		case 12:
			goto st294
		case 13:
			goto st104
		case 32:
			goto st446
		case 34:
			goto tr97
		case 44:
			goto st6
		case 61:
			goto st6
		case 92:
			goto tr163
		}
		goto tr160
tr690:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st447
	st447:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof447
		}
	st_case_447:
//line plugins/parsers/influx/machine.go:15839
		switch ( m.data)[( m.p)] {
		case 9:
			goto st446
		case 10:
			goto st317
		case 11:
			goto tr690
		case 12:
			goto st294
		case 13:
			goto st104
		case 32:
			goto st446
		case 34:
			goto tr97
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto tr163
		}
		goto tr160
tr692:
	( m.cs) = 448
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
tr687:
	( m.cs) = 448
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st448:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof448
		}
	st_case_448:
//line plugins/parsers/influx/machine.go:15908
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr691
		case 10:
			goto st317
		case 11:
			goto tr692
		case 12:
			goto tr556
		case 13:
			goto st104
		case 32:
			goto tr691
		case 34:
			goto tr204
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto tr205
		}
		goto tr202
	st449:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof449
		}
	st_case_449:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr686
		case 10:
			goto tr659
		case 11:
			goto tr687
		case 12:
			goto tr553
		case 13:
			goto tr661
		case 32:
			goto tr686
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st450
		}
		goto st67
	st450:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof450
		}
	st_case_450:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr686
		case 10:
			goto tr659
		case 11:
			goto tr687
		case 12:
			goto tr553
		case 13:
			goto tr661
		case 32:
			goto tr686
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st451
		}
		goto st67
	st451:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof451
		}
	st_case_451:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr686
		case 10:
			goto tr659
		case 11:
			goto tr687
		case 12:
			goto tr553
		case 13:
			goto tr661
		case 32:
			goto tr686
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st452
		}
		goto st67
	st452:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof452
		}
	st_case_452:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr686
		case 10:
			goto tr659
		case 11:
			goto tr687
		case 12:
			goto tr553
		case 13:
			goto tr661
		case 32:
			goto tr686
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st453
		}
		goto st67
	st453:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof453
		}
	st_case_453:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr686
		case 10:
			goto tr659
		case 11:
			goto tr687
		case 12:
			goto tr553
		case 13:
			goto tr661
		case 32:
			goto tr686
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st454
		}
		goto st67
	st454:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof454
		}
	st_case_454:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr686
		case 10:
			goto tr659
		case 11:
			goto tr687
		case 12:
			goto tr553
		case 13:
			goto tr661
		case 32:
			goto tr686
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st455
		}
		goto st67
	st455:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof455
		}
	st_case_455:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr686
		case 10:
			goto tr659
		case 11:
			goto tr687
		case 12:
			goto tr553
		case 13:
			goto tr661
		case 32:
			goto tr686
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st456
		}
		goto st67
	st456:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof456
		}
	st_case_456:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr686
		case 10:
			goto tr659
		case 11:
			goto tr687
		case 12:
			goto tr553
		case 13:
			goto tr661
		case 32:
			goto tr686
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st457
		}
		goto st67
	st457:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof457
		}
	st_case_457:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr686
		case 10:
			goto tr659
		case 11:
			goto tr687
		case 12:
			goto tr553
		case 13:
			goto tr661
		case 32:
			goto tr686
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st458
		}
		goto st67
	st458:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof458
		}
	st_case_458:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr686
		case 10:
			goto tr659
		case 11:
			goto tr687
		case 12:
			goto tr553
		case 13:
			goto tr661
		case 32:
			goto tr686
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st459
		}
		goto st67
	st459:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof459
		}
	st_case_459:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr686
		case 10:
			goto tr659
		case 11:
			goto tr687
		case 12:
			goto tr553
		case 13:
			goto tr661
		case 32:
			goto tr686
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st460
		}
		goto st67
	st460:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof460
		}
	st_case_460:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr686
		case 10:
			goto tr659
		case 11:
			goto tr687
		case 12:
			goto tr553
		case 13:
			goto tr661
		case 32:
			goto tr686
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st461
		}
		goto st67
	st461:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof461
		}
	st_case_461:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr686
		case 10:
			goto tr659
		case 11:
			goto tr687
		case 12:
			goto tr553
		case 13:
			goto tr661
		case 32:
			goto tr686
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st462
		}
		goto st67
	st462:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof462
		}
	st_case_462:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr686
		case 10:
			goto tr659
		case 11:
			goto tr687
		case 12:
			goto tr553
		case 13:
			goto tr661
		case 32:
			goto tr686
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st463
		}
		goto st67
	st463:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof463
		}
	st_case_463:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr686
		case 10:
			goto tr659
		case 11:
			goto tr687
		case 12:
			goto tr553
		case 13:
			goto tr661
		case 32:
			goto tr686
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st464
		}
		goto st67
	st464:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof464
		}
	st_case_464:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr686
		case 10:
			goto tr659
		case 11:
			goto tr687
		case 12:
			goto tr553
		case 13:
			goto tr661
		case 32:
			goto tr686
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st465
		}
		goto st67
	st465:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof465
		}
	st_case_465:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr686
		case 10:
			goto tr659
		case 11:
			goto tr687
		case 12:
			goto tr553
		case 13:
			goto tr661
		case 32:
			goto tr686
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st466
		}
		goto st67
	st466:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof466
		}
	st_case_466:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr686
		case 10:
			goto tr659
		case 11:
			goto tr687
		case 12:
			goto tr553
		case 13:
			goto tr661
		case 32:
			goto tr686
		case 34:
			goto tr208
		case 44:
			goto tr158
		case 61:
			goto tr165
		case 92:
			goto st69
		}
		goto st67
tr653:
	( m.cs) = 108
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
tr839:
	( m.cs) = 108
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr713:
	( m.cs) = 108
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr726:
	( m.cs) = 108
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr733:
	( m.cs) = 108
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr740:
	( m.cs) = 108
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr871:
	( m.cs) = 108
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr875:
	( m.cs) = 108
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr879:
	( m.cs) = 108
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st108:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof108
		}
	st_case_108:
//line plugins/parsers/influx/machine.go:16693
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr47
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr258
		case 44:
			goto st6
		case 61:
			goto st6
		case 92:
			goto tr279
		}
		goto tr278
tr278:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st109
	st109:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof109
		}
	st_case_109:
//line plugins/parsers/influx/machine.go:16726
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr47
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr261
		case 44:
			goto st6
		case 61:
			goto tr281
		case 92:
			goto st123
		}
		goto st109
tr281:
//line plugins/parsers/influx/machine.go.rl:86

	key = m.text()

//line plugins/parsers/influx/machine.go.rl:99

	key = m.text()

	goto st110
	st110:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof110
		}
	st_case_110:
//line plugins/parsers/influx/machine.go:16763
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr47
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr266
		case 44:
			goto st6
		case 45:
			goto tr283
		case 46:
			goto tr284
		case 48:
			goto tr285
		case 61:
			goto st6
		case 70:
			goto tr287
		case 84:
			goto tr288
		case 92:
			goto tr153
		case 102:
			goto tr289
		case 116:
			goto tr290
		}
		if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr286
		}
		goto tr148
tr283:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st111
	st111:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof111
		}
	st_case_111:
//line plugins/parsers/influx/machine.go:16813
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr155
		case 10:
			goto st7
		case 11:
			goto tr156
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr155
		case 34:
			goto tr157
		case 44:
			goto tr158
		case 46:
			goto st112
		case 48:
			goto st471
		case 61:
			goto st6
		case 92:
			goto st64
		}
		if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st474
		}
		goto st49
tr284:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st112
	st112:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof112
		}
	st_case_112:
//line plugins/parsers/influx/machine.go:16855
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr155
		case 10:
			goto st7
		case 11:
			goto tr156
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr155
		case 34:
			goto tr157
		case 44:
			goto tr158
		case 61:
			goto st6
		case 92:
			goto st64
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st467
		}
		goto st49
	st467:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof467
		}
	st_case_467:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr710
		case 10:
			goto tr515
		case 11:
			goto tr711
		case 12:
			goto tr712
		case 13:
			goto tr517
		case 32:
			goto tr710
		case 34:
			goto tr157
		case 44:
			goto tr713
		case 61:
			goto st6
		case 69:
			goto st113
		case 92:
			goto st64
		case 101:
			goto st113
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st467
		}
		goto st49
	st113:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof113
		}
	st_case_113:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr155
		case 10:
			goto st7
		case 11:
			goto tr156
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr155
		case 34:
			goto tr295
		case 44:
			goto tr158
		case 61:
			goto st6
		case 92:
			goto st64
		}
		switch {
		case ( m.data)[( m.p)] > 45:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st470
			}
		case ( m.data)[( m.p)] >= 43:
			goto st114
		}
		goto st49
tr295:
	( m.cs) = 468
//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st468:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof468
		}
	st_case_468:
//line plugins/parsers/influx/machine.go:16971
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto tr548
		case 13:
			goto st34
		case 32:
			goto tr547
		case 44:
			goto tr549
		case 61:
			goto tr132
		case 92:
			goto st23
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st469
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr547
		}
		goto st17
	st469:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof469
		}
	st_case_469:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 11:
			goto tr716
		case 13:
			goto tr717
		case 32:
			goto tr712
		case 44:
			goto tr718
		case 61:
			goto tr132
		case 92:
			goto st23
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st469
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr712
		}
		goto st17
	st114:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof114
		}
	st_case_114:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr155
		case 10:
			goto st7
		case 11:
			goto tr156
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr155
		case 34:
			goto tr157
		case 44:
			goto tr158
		case 61:
			goto st6
		case 92:
			goto st64
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st470
		}
		goto st49
	st470:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof470
		}
	st_case_470:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr710
		case 10:
			goto tr515
		case 11:
			goto tr711
		case 12:
			goto tr712
		case 13:
			goto tr517
		case 32:
			goto tr710
		case 34:
			goto tr157
		case 44:
			goto tr713
		case 61:
			goto st6
		case 92:
			goto st64
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st470
		}
		goto st49
	st471:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof471
		}
	st_case_471:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr710
		case 10:
			goto tr515
		case 11:
			goto tr711
		case 12:
			goto tr712
		case 13:
			goto tr517
		case 32:
			goto tr710
		case 34:
			goto tr157
		case 44:
			goto tr713
		case 46:
			goto st467
		case 61:
			goto st6
		case 69:
			goto st113
		case 92:
			goto st64
		case 101:
			goto st113
		case 105:
			goto st473
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st472
		}
		goto st49
	st472:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof472
		}
	st_case_472:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr710
		case 10:
			goto tr515
		case 11:
			goto tr711
		case 12:
			goto tr712
		case 13:
			goto tr517
		case 32:
			goto tr710
		case 34:
			goto tr157
		case 44:
			goto tr713
		case 46:
			goto st467
		case 61:
			goto st6
		case 69:
			goto st113
		case 92:
			goto st64
		case 101:
			goto st113
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st472
		}
		goto st49
	st473:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof473
		}
	st_case_473:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr721
		case 10:
			goto tr722
		case 11:
			goto tr723
		case 12:
			goto tr724
		case 13:
			goto tr725
		case 32:
			goto tr721
		case 34:
			goto tr157
		case 44:
			goto tr726
		case 61:
			goto st6
		case 92:
			goto st64
		}
		goto st49
	st474:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof474
		}
	st_case_474:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr710
		case 10:
			goto tr515
		case 11:
			goto tr711
		case 12:
			goto tr712
		case 13:
			goto tr517
		case 32:
			goto tr710
		case 34:
			goto tr157
		case 44:
			goto tr713
		case 46:
			goto st467
		case 61:
			goto st6
		case 69:
			goto st113
		case 92:
			goto st64
		case 101:
			goto st113
		case 105:
			goto st473
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st474
		}
		goto st49
tr285:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st475
	st475:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof475
		}
	st_case_475:
//line plugins/parsers/influx/machine.go:17243
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr710
		case 10:
			goto tr515
		case 11:
			goto tr711
		case 12:
			goto tr712
		case 13:
			goto tr517
		case 32:
			goto tr710
		case 34:
			goto tr157
		case 44:
			goto tr713
		case 46:
			goto st467
		case 61:
			goto st6
		case 69:
			goto st113
		case 92:
			goto st64
		case 101:
			goto st113
		case 105:
			goto st473
		case 117:
			goto st476
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st472
		}
		goto st49
	st476:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof476
		}
	st_case_476:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr728
		case 10:
			goto tr729
		case 11:
			goto tr730
		case 12:
			goto tr731
		case 13:
			goto tr732
		case 32:
			goto tr728
		case 34:
			goto tr157
		case 44:
			goto tr733
		case 61:
			goto st6
		case 92:
			goto st64
		}
		goto st49
tr286:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st477
	st477:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof477
		}
	st_case_477:
//line plugins/parsers/influx/machine.go:17319
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr710
		case 10:
			goto tr515
		case 11:
			goto tr711
		case 12:
			goto tr712
		case 13:
			goto tr517
		case 32:
			goto tr710
		case 34:
			goto tr157
		case 44:
			goto tr713
		case 46:
			goto st467
		case 61:
			goto st6
		case 69:
			goto st113
		case 92:
			goto st64
		case 101:
			goto st113
		case 105:
			goto st473
		case 117:
			goto st476
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st477
		}
		goto st49
tr287:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st478
	st478:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof478
		}
	st_case_478:
//line plugins/parsers/influx/machine.go:17367
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr735
		case 10:
			goto tr736
		case 11:
			goto tr737
		case 12:
			goto tr738
		case 13:
			goto tr739
		case 32:
			goto tr735
		case 34:
			goto tr157
		case 44:
			goto tr740
		case 61:
			goto st6
		case 65:
			goto st115
		case 92:
			goto st64
		case 97:
			goto st118
		}
		goto st49
	st115:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof115
		}
	st_case_115:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr155
		case 10:
			goto st7
		case 11:
			goto tr156
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr155
		case 34:
			goto tr157
		case 44:
			goto tr158
		case 61:
			goto st6
		case 76:
			goto st116
		case 92:
			goto st64
		}
		goto st49
	st116:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof116
		}
	st_case_116:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr155
		case 10:
			goto st7
		case 11:
			goto tr156
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr155
		case 34:
			goto tr157
		case 44:
			goto tr158
		case 61:
			goto st6
		case 83:
			goto st117
		case 92:
			goto st64
		}
		goto st49
	st117:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof117
		}
	st_case_117:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr155
		case 10:
			goto st7
		case 11:
			goto tr156
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr155
		case 34:
			goto tr157
		case 44:
			goto tr158
		case 61:
			goto st6
		case 69:
			goto st479
		case 92:
			goto st64
		}
		goto st49
	st479:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof479
		}
	st_case_479:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr735
		case 10:
			goto tr736
		case 11:
			goto tr737
		case 12:
			goto tr738
		case 13:
			goto tr739
		case 32:
			goto tr735
		case 34:
			goto tr157
		case 44:
			goto tr740
		case 61:
			goto st6
		case 92:
			goto st64
		}
		goto st49
	st118:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof118
		}
	st_case_118:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr155
		case 10:
			goto st7
		case 11:
			goto tr156
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr155
		case 34:
			goto tr157
		case 44:
			goto tr158
		case 61:
			goto st6
		case 92:
			goto st64
		case 108:
			goto st119
		}
		goto st49
	st119:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof119
		}
	st_case_119:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr155
		case 10:
			goto st7
		case 11:
			goto tr156
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr155
		case 34:
			goto tr157
		case 44:
			goto tr158
		case 61:
			goto st6
		case 92:
			goto st64
		case 115:
			goto st120
		}
		goto st49
	st120:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof120
		}
	st_case_120:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr155
		case 10:
			goto st7
		case 11:
			goto tr156
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr155
		case 34:
			goto tr157
		case 44:
			goto tr158
		case 61:
			goto st6
		case 92:
			goto st64
		case 101:
			goto st479
		}
		goto st49
tr288:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st480
	st480:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof480
		}
	st_case_480:
//line plugins/parsers/influx/machine.go:17614
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr735
		case 10:
			goto tr736
		case 11:
			goto tr737
		case 12:
			goto tr738
		case 13:
			goto tr739
		case 32:
			goto tr735
		case 34:
			goto tr157
		case 44:
			goto tr740
		case 61:
			goto st6
		case 82:
			goto st121
		case 92:
			goto st64
		case 114:
			goto st122
		}
		goto st49
	st121:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof121
		}
	st_case_121:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr155
		case 10:
			goto st7
		case 11:
			goto tr156
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr155
		case 34:
			goto tr157
		case 44:
			goto tr158
		case 61:
			goto st6
		case 85:
			goto st117
		case 92:
			goto st64
		}
		goto st49
	st122:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof122
		}
	st_case_122:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr155
		case 10:
			goto st7
		case 11:
			goto tr156
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr155
		case 34:
			goto tr157
		case 44:
			goto tr158
		case 61:
			goto st6
		case 92:
			goto st64
		case 117:
			goto st120
		}
		goto st49
tr289:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st481
	st481:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof481
		}
	st_case_481:
//line plugins/parsers/influx/machine.go:17713
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr735
		case 10:
			goto tr736
		case 11:
			goto tr737
		case 12:
			goto tr738
		case 13:
			goto tr739
		case 32:
			goto tr735
		case 34:
			goto tr157
		case 44:
			goto tr740
		case 61:
			goto st6
		case 92:
			goto st64
		case 97:
			goto st118
		}
		goto st49
tr290:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st482
	st482:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof482
		}
	st_case_482:
//line plugins/parsers/influx/machine.go:17750
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr735
		case 10:
			goto tr736
		case 11:
			goto tr737
		case 12:
			goto tr738
		case 13:
			goto tr739
		case 32:
			goto tr735
		case 34:
			goto tr157
		case 44:
			goto tr740
		case 61:
			goto st6
		case 92:
			goto st64
		case 114:
			goto st122
		}
		goto st49
tr279:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st123
	st123:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof123
		}
	st_case_123:
//line plugins/parsers/influx/machine.go:17787
		switch ( m.data)[( m.p)] {
		case 34:
			goto st109
		case 92:
			goto st124
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr47
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr47
		}
		goto st46
	st124:
//line plugins/parsers/influx/machine.go.rl:234
 ( m.p)--
 
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof124
		}
	st_case_124:
//line plugins/parsers/influx/machine.go:17811
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr47
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr261
		case 44:
			goto st6
		case 61:
			goto tr281
		case 92:
			goto st123
		}
		goto st109
tr267:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st125
	st125:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof125
		}
	st_case_125:
//line plugins/parsers/influx/machine.go:17844
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr231
		case 10:
			goto st7
		case 11:
			goto tr232
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr231
		case 34:
			goto tr157
		case 44:
			goto tr233
		case 46:
			goto st126
		case 48:
			goto st507
		case 61:
			goto st6
		case 92:
			goto st87
		}
		if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st510
		}
		goto st81
tr268:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st126
	st126:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof126
		}
	st_case_126:
//line plugins/parsers/influx/machine.go:17886
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr231
		case 10:
			goto st7
		case 11:
			goto tr232
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr231
		case 34:
			goto tr157
		case 44:
			goto tr233
		case 61:
			goto st6
		case 92:
			goto st87
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st483
		}
		goto st81
	st483:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof483
		}
	st_case_483:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr745
		case 10:
			goto tr620
		case 11:
			goto tr746
		case 12:
			goto tr712
		case 13:
			goto tr623
		case 32:
			goto tr745
		case 34:
			goto tr157
		case 44:
			goto tr747
		case 61:
			goto st6
		case 69:
			goto st128
		case 92:
			goto st87
		case 101:
			goto st128
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st483
		}
		goto st81
tr746:
	( m.cs) = 484
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr779:
	( m.cs) = 484
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr785:
	( m.cs) = 484
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr791:
	( m.cs) = 484
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st484:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof484
		}
	st_case_484:
//line plugins/parsers/influx/machine.go:18045
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr749
		case 10:
			goto st288
		case 11:
			goto tr750
		case 12:
			goto tr547
		case 13:
			goto st74
		case 32:
			goto tr749
		case 34:
			goto tr204
		case 44:
			goto tr233
		case 45:
			goto tr751
		case 61:
			goto st6
		case 92:
			goto tr237
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr752
		}
		goto tr235
tr750:
	( m.cs) = 485
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
	st485:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof485
		}
	st_case_485:
//line plugins/parsers/influx/machine.go:18096
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr749
		case 10:
			goto st288
		case 11:
			goto tr750
		case 12:
			goto tr547
		case 13:
			goto st74
		case 32:
			goto tr749
		case 34:
			goto tr204
		case 44:
			goto tr233
		case 45:
			goto tr751
		case 61:
			goto tr101
		case 92:
			goto tr237
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr752
		}
		goto tr235
tr751:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st127
	st127:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof127
		}
	st_case_127:
//line plugins/parsers/influx/machine.go:18136
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr231
		case 10:
			goto st7
		case 11:
			goto tr239
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr231
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st486
		}
		goto st83
tr752:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st486
	st486:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof486
		}
	st_case_486:
//line plugins/parsers/influx/machine.go:18174
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr753
		case 10:
			goto tr584
		case 11:
			goto tr754
		case 12:
			goto tr553
		case 13:
			goto tr586
		case 32:
			goto tr753
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st488
		}
		goto st83
tr757:
	( m.cs) = 487
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
tr754:
	( m.cs) = 487
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st487:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof487
		}
	st_case_487:
//line plugins/parsers/influx/machine.go:18246
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr756
		case 10:
			goto st288
		case 11:
			goto tr757
		case 12:
			goto tr556
		case 13:
			goto st74
		case 32:
			goto tr756
		case 34:
			goto tr204
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto tr237
		}
		goto tr235
	st488:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof488
		}
	st_case_488:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr753
		case 10:
			goto tr584
		case 11:
			goto tr754
		case 12:
			goto tr553
		case 13:
			goto tr586
		case 32:
			goto tr753
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st489
		}
		goto st83
	st489:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof489
		}
	st_case_489:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr753
		case 10:
			goto tr584
		case 11:
			goto tr754
		case 12:
			goto tr553
		case 13:
			goto tr586
		case 32:
			goto tr753
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st490
		}
		goto st83
	st490:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof490
		}
	st_case_490:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr753
		case 10:
			goto tr584
		case 11:
			goto tr754
		case 12:
			goto tr553
		case 13:
			goto tr586
		case 32:
			goto tr753
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st491
		}
		goto st83
	st491:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof491
		}
	st_case_491:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr753
		case 10:
			goto tr584
		case 11:
			goto tr754
		case 12:
			goto tr553
		case 13:
			goto tr586
		case 32:
			goto tr753
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st492
		}
		goto st83
	st492:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof492
		}
	st_case_492:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr753
		case 10:
			goto tr584
		case 11:
			goto tr754
		case 12:
			goto tr553
		case 13:
			goto tr586
		case 32:
			goto tr753
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st493
		}
		goto st83
	st493:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof493
		}
	st_case_493:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr753
		case 10:
			goto tr584
		case 11:
			goto tr754
		case 12:
			goto tr553
		case 13:
			goto tr586
		case 32:
			goto tr753
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st494
		}
		goto st83
	st494:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof494
		}
	st_case_494:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr753
		case 10:
			goto tr584
		case 11:
			goto tr754
		case 12:
			goto tr553
		case 13:
			goto tr586
		case 32:
			goto tr753
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st495
		}
		goto st83
	st495:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof495
		}
	st_case_495:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr753
		case 10:
			goto tr584
		case 11:
			goto tr754
		case 12:
			goto tr553
		case 13:
			goto tr586
		case 32:
			goto tr753
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st496
		}
		goto st83
	st496:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof496
		}
	st_case_496:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr753
		case 10:
			goto tr584
		case 11:
			goto tr754
		case 12:
			goto tr553
		case 13:
			goto tr586
		case 32:
			goto tr753
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st497
		}
		goto st83
	st497:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof497
		}
	st_case_497:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr753
		case 10:
			goto tr584
		case 11:
			goto tr754
		case 12:
			goto tr553
		case 13:
			goto tr586
		case 32:
			goto tr753
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st498
		}
		goto st83
	st498:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof498
		}
	st_case_498:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr753
		case 10:
			goto tr584
		case 11:
			goto tr754
		case 12:
			goto tr553
		case 13:
			goto tr586
		case 32:
			goto tr753
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st499
		}
		goto st83
	st499:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof499
		}
	st_case_499:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr753
		case 10:
			goto tr584
		case 11:
			goto tr754
		case 12:
			goto tr553
		case 13:
			goto tr586
		case 32:
			goto tr753
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st500
		}
		goto st83
	st500:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof500
		}
	st_case_500:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr753
		case 10:
			goto tr584
		case 11:
			goto tr754
		case 12:
			goto tr553
		case 13:
			goto tr586
		case 32:
			goto tr753
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st501
		}
		goto st83
	st501:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof501
		}
	st_case_501:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr753
		case 10:
			goto tr584
		case 11:
			goto tr754
		case 12:
			goto tr553
		case 13:
			goto tr586
		case 32:
			goto tr753
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st502
		}
		goto st83
	st502:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof502
		}
	st_case_502:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr753
		case 10:
			goto tr584
		case 11:
			goto tr754
		case 12:
			goto tr553
		case 13:
			goto tr586
		case 32:
			goto tr753
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st503
		}
		goto st83
	st503:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof503
		}
	st_case_503:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr753
		case 10:
			goto tr584
		case 11:
			goto tr754
		case 12:
			goto tr553
		case 13:
			goto tr586
		case 32:
			goto tr753
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st504
		}
		goto st83
	st504:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof504
		}
	st_case_504:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr753
		case 10:
			goto tr584
		case 11:
			goto tr754
		case 12:
			goto tr553
		case 13:
			goto tr586
		case 32:
			goto tr753
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st505
		}
		goto st83
	st505:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof505
		}
	st_case_505:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr753
		case 10:
			goto tr584
		case 11:
			goto tr754
		case 12:
			goto tr553
		case 13:
			goto tr586
		case 32:
			goto tr753
		case 34:
			goto tr208
		case 44:
			goto tr233
		case 61:
			goto tr101
		case 92:
			goto st85
		}
		goto st83
	st128:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof128
		}
	st_case_128:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr231
		case 10:
			goto st7
		case 11:
			goto tr232
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr231
		case 34:
			goto tr295
		case 44:
			goto tr233
		case 61:
			goto st6
		case 92:
			goto st87
		}
		switch {
		case ( m.data)[( m.p)] > 45:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st506
			}
		case ( m.data)[( m.p)] >= 43:
			goto st129
		}
		goto st81
	st129:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof129
		}
	st_case_129:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr231
		case 10:
			goto st7
		case 11:
			goto tr232
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr231
		case 34:
			goto tr157
		case 44:
			goto tr233
		case 61:
			goto st6
		case 92:
			goto st87
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st506
		}
		goto st81
	st506:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof506
		}
	st_case_506:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr745
		case 10:
			goto tr620
		case 11:
			goto tr746
		case 12:
			goto tr712
		case 13:
			goto tr623
		case 32:
			goto tr745
		case 34:
			goto tr157
		case 44:
			goto tr747
		case 61:
			goto st6
		case 92:
			goto st87
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st506
		}
		goto st81
	st507:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof507
		}
	st_case_507:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr745
		case 10:
			goto tr620
		case 11:
			goto tr746
		case 12:
			goto tr712
		case 13:
			goto tr623
		case 32:
			goto tr745
		case 34:
			goto tr157
		case 44:
			goto tr747
		case 46:
			goto st483
		case 61:
			goto st6
		case 69:
			goto st128
		case 92:
			goto st87
		case 101:
			goto st128
		case 105:
			goto st509
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st508
		}
		goto st81
	st508:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof508
		}
	st_case_508:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr745
		case 10:
			goto tr620
		case 11:
			goto tr746
		case 12:
			goto tr712
		case 13:
			goto tr623
		case 32:
			goto tr745
		case 34:
			goto tr157
		case 44:
			goto tr747
		case 46:
			goto st483
		case 61:
			goto st6
		case 69:
			goto st128
		case 92:
			goto st87
		case 101:
			goto st128
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st508
		}
		goto st81
	st509:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof509
		}
	st_case_509:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr777
		case 10:
			goto tr778
		case 11:
			goto tr779
		case 12:
			goto tr724
		case 13:
			goto tr780
		case 32:
			goto tr777
		case 34:
			goto tr157
		case 44:
			goto tr781
		case 61:
			goto st6
		case 92:
			goto st87
		}
		goto st81
	st510:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof510
		}
	st_case_510:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr745
		case 10:
			goto tr620
		case 11:
			goto tr746
		case 12:
			goto tr712
		case 13:
			goto tr623
		case 32:
			goto tr745
		case 34:
			goto tr157
		case 44:
			goto tr747
		case 46:
			goto st483
		case 61:
			goto st6
		case 69:
			goto st128
		case 92:
			goto st87
		case 101:
			goto st128
		case 105:
			goto st509
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st510
		}
		goto st81
tr269:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st511
	st511:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof511
		}
	st_case_511:
//line plugins/parsers/influx/machine.go:19077
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr745
		case 10:
			goto tr620
		case 11:
			goto tr746
		case 12:
			goto tr712
		case 13:
			goto tr623
		case 32:
			goto tr745
		case 34:
			goto tr157
		case 44:
			goto tr747
		case 46:
			goto st483
		case 61:
			goto st6
		case 69:
			goto st128
		case 92:
			goto st87
		case 101:
			goto st128
		case 105:
			goto st509
		case 117:
			goto st512
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st508
		}
		goto st81
	st512:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof512
		}
	st_case_512:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr783
		case 10:
			goto tr784
		case 11:
			goto tr785
		case 12:
			goto tr731
		case 13:
			goto tr786
		case 32:
			goto tr783
		case 34:
			goto tr157
		case 44:
			goto tr787
		case 61:
			goto st6
		case 92:
			goto st87
		}
		goto st81
tr270:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st513
	st513:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof513
		}
	st_case_513:
//line plugins/parsers/influx/machine.go:19153
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr745
		case 10:
			goto tr620
		case 11:
			goto tr746
		case 12:
			goto tr712
		case 13:
			goto tr623
		case 32:
			goto tr745
		case 34:
			goto tr157
		case 44:
			goto tr747
		case 46:
			goto st483
		case 61:
			goto st6
		case 69:
			goto st128
		case 92:
			goto st87
		case 101:
			goto st128
		case 105:
			goto st509
		case 117:
			goto st512
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st513
		}
		goto st81
tr271:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st514
	st514:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof514
		}
	st_case_514:
//line plugins/parsers/influx/machine.go:19201
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr789
		case 10:
			goto tr790
		case 11:
			goto tr791
		case 12:
			goto tr738
		case 13:
			goto tr792
		case 32:
			goto tr789
		case 34:
			goto tr157
		case 44:
			goto tr793
		case 61:
			goto st6
		case 65:
			goto st130
		case 92:
			goto st87
		case 97:
			goto st133
		}
		goto st81
	st130:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof130
		}
	st_case_130:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr231
		case 10:
			goto st7
		case 11:
			goto tr232
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr231
		case 34:
			goto tr157
		case 44:
			goto tr233
		case 61:
			goto st6
		case 76:
			goto st131
		case 92:
			goto st87
		}
		goto st81
	st131:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof131
		}
	st_case_131:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr231
		case 10:
			goto st7
		case 11:
			goto tr232
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr231
		case 34:
			goto tr157
		case 44:
			goto tr233
		case 61:
			goto st6
		case 83:
			goto st132
		case 92:
			goto st87
		}
		goto st81
	st132:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof132
		}
	st_case_132:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr231
		case 10:
			goto st7
		case 11:
			goto tr232
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr231
		case 34:
			goto tr157
		case 44:
			goto tr233
		case 61:
			goto st6
		case 69:
			goto st515
		case 92:
			goto st87
		}
		goto st81
	st515:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof515
		}
	st_case_515:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr789
		case 10:
			goto tr790
		case 11:
			goto tr791
		case 12:
			goto tr738
		case 13:
			goto tr792
		case 32:
			goto tr789
		case 34:
			goto tr157
		case 44:
			goto tr793
		case 61:
			goto st6
		case 92:
			goto st87
		}
		goto st81
	st133:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof133
		}
	st_case_133:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr231
		case 10:
			goto st7
		case 11:
			goto tr232
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr231
		case 34:
			goto tr157
		case 44:
			goto tr233
		case 61:
			goto st6
		case 92:
			goto st87
		case 108:
			goto st134
		}
		goto st81
	st134:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof134
		}
	st_case_134:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr231
		case 10:
			goto st7
		case 11:
			goto tr232
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr231
		case 34:
			goto tr157
		case 44:
			goto tr233
		case 61:
			goto st6
		case 92:
			goto st87
		case 115:
			goto st135
		}
		goto st81
	st135:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof135
		}
	st_case_135:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr231
		case 10:
			goto st7
		case 11:
			goto tr232
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr231
		case 34:
			goto tr157
		case 44:
			goto tr233
		case 61:
			goto st6
		case 92:
			goto st87
		case 101:
			goto st515
		}
		goto st81
tr272:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st516
	st516:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof516
		}
	st_case_516:
//line plugins/parsers/influx/machine.go:19448
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr789
		case 10:
			goto tr790
		case 11:
			goto tr791
		case 12:
			goto tr738
		case 13:
			goto tr792
		case 32:
			goto tr789
		case 34:
			goto tr157
		case 44:
			goto tr793
		case 61:
			goto st6
		case 82:
			goto st136
		case 92:
			goto st87
		case 114:
			goto st137
		}
		goto st81
	st136:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof136
		}
	st_case_136:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr231
		case 10:
			goto st7
		case 11:
			goto tr232
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr231
		case 34:
			goto tr157
		case 44:
			goto tr233
		case 61:
			goto st6
		case 85:
			goto st132
		case 92:
			goto st87
		}
		goto st81
	st137:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof137
		}
	st_case_137:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr231
		case 10:
			goto st7
		case 11:
			goto tr232
		case 12:
			goto tr60
		case 13:
			goto st8
		case 32:
			goto tr231
		case 34:
			goto tr157
		case 44:
			goto tr233
		case 61:
			goto st6
		case 92:
			goto st87
		case 117:
			goto st135
		}
		goto st81
tr273:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st517
	st517:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof517
		}
	st_case_517:
//line plugins/parsers/influx/machine.go:19547
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr789
		case 10:
			goto tr790
		case 11:
			goto tr791
		case 12:
			goto tr738
		case 13:
			goto tr792
		case 32:
			goto tr789
		case 34:
			goto tr157
		case 44:
			goto tr793
		case 61:
			goto st6
		case 92:
			goto st87
		case 97:
			goto st133
		}
		goto st81
tr274:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st518
	st518:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof518
		}
	st_case_518:
//line plugins/parsers/influx/machine.go:19584
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr789
		case 10:
			goto tr790
		case 11:
			goto tr791
		case 12:
			goto tr738
		case 13:
			goto tr792
		case 32:
			goto tr789
		case 34:
			goto tr157
		case 44:
			goto tr793
		case 61:
			goto st6
		case 92:
			goto st87
		case 114:
			goto st137
		}
		goto st81
tr259:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st138
	st138:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof138
		}
	st_case_138:
//line plugins/parsers/influx/machine.go:19621
		switch ( m.data)[( m.p)] {
		case 34:
			goto st99
		case 92:
			goto st139
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr47
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr47
		}
		goto st46
	st139:
//line plugins/parsers/influx/machine.go.rl:234
 ( m.p)--
 
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof139
		}
	st_case_139:
//line plugins/parsers/influx/machine.go:19645
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr47
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr261
		case 44:
			goto st6
		case 61:
			goto tr262
		case 92:
			goto st138
		}
		goto st99
	st140:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof140
		}
	st_case_140:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr89
		case 10:
			goto st7
		case 11:
			goto tr90
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr89
		case 34:
			goto tr317
		case 44:
			goto tr92
		case 92:
			goto st142
		}
		switch {
		case ( m.data)[( m.p)] > 45:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st521
			}
		case ( m.data)[( m.p)] >= 43:
			goto st141
		}
		goto st31
tr317:
	( m.cs) = 519
//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st519:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof519
		}
	st_case_519:
//line plugins/parsers/influx/machine.go:19719
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 11:
			goto tr618
		case 13:
			goto st34
		case 32:
			goto tr482
		case 44:
			goto tr484
		case 92:
			goto st96
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st520
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr482
		}
		goto st1
	st520:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof520
		}
	st_case_520:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 11:
			goto tr798
		case 13:
			goto tr717
		case 32:
			goto tr622
		case 44:
			goto tr799
		case 92:
			goto st96
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st520
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr622
		}
		goto st1
	st141:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof141
		}
	st_case_141:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr89
		case 10:
			goto st7
		case 11:
			goto tr90
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr89
		case 34:
			goto tr91
		case 44:
			goto tr92
		case 92:
			goto st142
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st521
		}
		goto st31
	st521:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof521
		}
	st_case_521:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr619
		case 10:
			goto tr620
		case 11:
			goto tr621
		case 12:
			goto tr622
		case 13:
			goto tr623
		case 32:
			goto tr619
		case 34:
			goto tr91
		case 44:
			goto tr624
		case 92:
			goto st142
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st521
		}
		goto st31
tr87:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st142
	st142:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof142
		}
	st_case_142:
//line plugins/parsers/influx/machine.go:19840
		switch ( m.data)[( m.p)] {
		case 34:
			goto st31
		case 92:
			goto st31
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr8
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr8
		}
		goto st1
	st522:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof522
		}
	st_case_522:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr619
		case 10:
			goto tr620
		case 11:
			goto tr621
		case 12:
			goto tr622
		case 13:
			goto tr623
		case 32:
			goto tr619
		case 34:
			goto tr91
		case 44:
			goto tr624
		case 46:
			goto st396
		case 69:
			goto st140
		case 92:
			goto st142
		case 101:
			goto st140
		case 105:
			goto st524
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st523
		}
		goto st31
	st523:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof523
		}
	st_case_523:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr619
		case 10:
			goto tr620
		case 11:
			goto tr621
		case 12:
			goto tr622
		case 13:
			goto tr623
		case 32:
			goto tr619
		case 34:
			goto tr91
		case 44:
			goto tr624
		case 46:
			goto st396
		case 69:
			goto st140
		case 92:
			goto st142
		case 101:
			goto st140
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st523
		}
		goto st31
	st524:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof524
		}
	st_case_524:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr802
		case 10:
			goto tr778
		case 11:
			goto tr803
		case 12:
			goto tr804
		case 13:
			goto tr780
		case 32:
			goto tr802
		case 34:
			goto tr91
		case 44:
			goto tr805
		case 92:
			goto st142
		}
		goto st31
	st525:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof525
		}
	st_case_525:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr619
		case 10:
			goto tr620
		case 11:
			goto tr621
		case 12:
			goto tr622
		case 13:
			goto tr623
		case 32:
			goto tr619
		case 34:
			goto tr91
		case 44:
			goto tr624
		case 46:
			goto st396
		case 69:
			goto st140
		case 92:
			goto st142
		case 101:
			goto st140
		case 105:
			goto st524
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st525
		}
		goto st31
tr247:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st526
	st526:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof526
		}
	st_case_526:
//line plugins/parsers/influx/machine.go:20002
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr619
		case 10:
			goto tr620
		case 11:
			goto tr621
		case 12:
			goto tr622
		case 13:
			goto tr623
		case 32:
			goto tr619
		case 34:
			goto tr91
		case 44:
			goto tr624
		case 46:
			goto st396
		case 69:
			goto st140
		case 92:
			goto st142
		case 101:
			goto st140
		case 105:
			goto st524
		case 117:
			goto st527
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st523
		}
		goto st31
	st527:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof527
		}
	st_case_527:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr807
		case 10:
			goto tr784
		case 11:
			goto tr808
		case 12:
			goto tr809
		case 13:
			goto tr786
		case 32:
			goto tr807
		case 34:
			goto tr91
		case 44:
			goto tr810
		case 92:
			goto st142
		}
		goto st31
tr248:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st528
	st528:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof528
		}
	st_case_528:
//line plugins/parsers/influx/machine.go:20074
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr619
		case 10:
			goto tr620
		case 11:
			goto tr621
		case 12:
			goto tr622
		case 13:
			goto tr623
		case 32:
			goto tr619
		case 34:
			goto tr91
		case 44:
			goto tr624
		case 46:
			goto st396
		case 69:
			goto st140
		case 92:
			goto st142
		case 101:
			goto st140
		case 105:
			goto st524
		case 117:
			goto st527
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st528
		}
		goto st31
tr249:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st529
	st529:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof529
		}
	st_case_529:
//line plugins/parsers/influx/machine.go:20120
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr812
		case 10:
			goto tr790
		case 11:
			goto tr813
		case 12:
			goto tr814
		case 13:
			goto tr792
		case 32:
			goto tr812
		case 34:
			goto tr91
		case 44:
			goto tr815
		case 65:
			goto st143
		case 92:
			goto st142
		case 97:
			goto st146
		}
		goto st31
	st143:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof143
		}
	st_case_143:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr89
		case 10:
			goto st7
		case 11:
			goto tr90
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr89
		case 34:
			goto tr91
		case 44:
			goto tr92
		case 76:
			goto st144
		case 92:
			goto st142
		}
		goto st31
	st144:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof144
		}
	st_case_144:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr89
		case 10:
			goto st7
		case 11:
			goto tr90
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr89
		case 34:
			goto tr91
		case 44:
			goto tr92
		case 83:
			goto st145
		case 92:
			goto st142
		}
		goto st31
	st145:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof145
		}
	st_case_145:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr89
		case 10:
			goto st7
		case 11:
			goto tr90
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr89
		case 34:
			goto tr91
		case 44:
			goto tr92
		case 69:
			goto st530
		case 92:
			goto st142
		}
		goto st31
	st530:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof530
		}
	st_case_530:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr812
		case 10:
			goto tr790
		case 11:
			goto tr813
		case 12:
			goto tr814
		case 13:
			goto tr792
		case 32:
			goto tr812
		case 34:
			goto tr91
		case 44:
			goto tr815
		case 92:
			goto st142
		}
		goto st31
	st146:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof146
		}
	st_case_146:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr89
		case 10:
			goto st7
		case 11:
			goto tr90
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr89
		case 34:
			goto tr91
		case 44:
			goto tr92
		case 92:
			goto st142
		case 108:
			goto st147
		}
		goto st31
	st147:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof147
		}
	st_case_147:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr89
		case 10:
			goto st7
		case 11:
			goto tr90
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr89
		case 34:
			goto tr91
		case 44:
			goto tr92
		case 92:
			goto st142
		case 115:
			goto st148
		}
		goto st31
	st148:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof148
		}
	st_case_148:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr89
		case 10:
			goto st7
		case 11:
			goto tr90
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr89
		case 34:
			goto tr91
		case 44:
			goto tr92
		case 92:
			goto st142
		case 101:
			goto st530
		}
		goto st31
tr250:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st531
	st531:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof531
		}
	st_case_531:
//line plugins/parsers/influx/machine.go:20351
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr812
		case 10:
			goto tr790
		case 11:
			goto tr813
		case 12:
			goto tr814
		case 13:
			goto tr792
		case 32:
			goto tr812
		case 34:
			goto tr91
		case 44:
			goto tr815
		case 82:
			goto st149
		case 92:
			goto st142
		case 114:
			goto st150
		}
		goto st31
	st149:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof149
		}
	st_case_149:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr89
		case 10:
			goto st7
		case 11:
			goto tr90
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr89
		case 34:
			goto tr91
		case 44:
			goto tr92
		case 85:
			goto st145
		case 92:
			goto st142
		}
		goto st31
	st150:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof150
		}
	st_case_150:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr89
		case 10:
			goto st7
		case 11:
			goto tr90
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr89
		case 34:
			goto tr91
		case 44:
			goto tr92
		case 92:
			goto st142
		case 117:
			goto st148
		}
		goto st31
tr251:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st532
	st532:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof532
		}
	st_case_532:
//line plugins/parsers/influx/machine.go:20444
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr812
		case 10:
			goto tr790
		case 11:
			goto tr813
		case 12:
			goto tr814
		case 13:
			goto tr792
		case 32:
			goto tr812
		case 34:
			goto tr91
		case 44:
			goto tr815
		case 92:
			goto st142
		case 97:
			goto st146
		}
		goto st31
tr252:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st533
	st533:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof533
		}
	st_case_533:
//line plugins/parsers/influx/machine.go:20479
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr812
		case 10:
			goto tr790
		case 11:
			goto tr813
		case 12:
			goto tr814
		case 13:
			goto tr792
		case 32:
			goto tr812
		case 34:
			goto tr91
		case 44:
			goto tr815
		case 92:
			goto st142
		case 114:
			goto st150
		}
		goto st31
	st534:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof534
		}
	st_case_534:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr611
		case 10:
			goto tr584
		case 11:
			goto tr612
		case 12:
			goto tr490
		case 13:
			goto tr586
		case 32:
			goto tr611
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st535
		}
		goto st42
	st535:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof535
		}
	st_case_535:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr611
		case 10:
			goto tr584
		case 11:
			goto tr612
		case 12:
			goto tr490
		case 13:
			goto tr586
		case 32:
			goto tr611
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st536
		}
		goto st42
	st536:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof536
		}
	st_case_536:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr611
		case 10:
			goto tr584
		case 11:
			goto tr612
		case 12:
			goto tr490
		case 13:
			goto tr586
		case 32:
			goto tr611
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st537
		}
		goto st42
	st537:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof537
		}
	st_case_537:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr611
		case 10:
			goto tr584
		case 11:
			goto tr612
		case 12:
			goto tr490
		case 13:
			goto tr586
		case 32:
			goto tr611
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st538
		}
		goto st42
	st538:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof538
		}
	st_case_538:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr611
		case 10:
			goto tr584
		case 11:
			goto tr612
		case 12:
			goto tr490
		case 13:
			goto tr586
		case 32:
			goto tr611
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st539
		}
		goto st42
	st539:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof539
		}
	st_case_539:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr611
		case 10:
			goto tr584
		case 11:
			goto tr612
		case 12:
			goto tr490
		case 13:
			goto tr586
		case 32:
			goto tr611
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st540
		}
		goto st42
	st540:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof540
		}
	st_case_540:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr611
		case 10:
			goto tr584
		case 11:
			goto tr612
		case 12:
			goto tr490
		case 13:
			goto tr586
		case 32:
			goto tr611
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st541
		}
		goto st42
	st541:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof541
		}
	st_case_541:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr611
		case 10:
			goto tr584
		case 11:
			goto tr612
		case 12:
			goto tr490
		case 13:
			goto tr586
		case 32:
			goto tr611
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st542
		}
		goto st42
	st542:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof542
		}
	st_case_542:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr611
		case 10:
			goto tr584
		case 11:
			goto tr612
		case 12:
			goto tr490
		case 13:
			goto tr586
		case 32:
			goto tr611
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st543
		}
		goto st42
	st543:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof543
		}
	st_case_543:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr611
		case 10:
			goto tr584
		case 11:
			goto tr612
		case 12:
			goto tr490
		case 13:
			goto tr586
		case 32:
			goto tr611
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st544
		}
		goto st42
	st544:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof544
		}
	st_case_544:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr611
		case 10:
			goto tr584
		case 11:
			goto tr612
		case 12:
			goto tr490
		case 13:
			goto tr586
		case 32:
			goto tr611
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st545
		}
		goto st42
	st545:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof545
		}
	st_case_545:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr611
		case 10:
			goto tr584
		case 11:
			goto tr612
		case 12:
			goto tr490
		case 13:
			goto tr586
		case 32:
			goto tr611
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st546
		}
		goto st42
	st546:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof546
		}
	st_case_546:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr611
		case 10:
			goto tr584
		case 11:
			goto tr612
		case 12:
			goto tr490
		case 13:
			goto tr586
		case 32:
			goto tr611
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st547
		}
		goto st42
	st547:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof547
		}
	st_case_547:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr611
		case 10:
			goto tr584
		case 11:
			goto tr612
		case 12:
			goto tr490
		case 13:
			goto tr586
		case 32:
			goto tr611
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st548
		}
		goto st42
	st548:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof548
		}
	st_case_548:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr611
		case 10:
			goto tr584
		case 11:
			goto tr612
		case 12:
			goto tr490
		case 13:
			goto tr586
		case 32:
			goto tr611
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st549
		}
		goto st42
	st549:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof549
		}
	st_case_549:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr611
		case 10:
			goto tr584
		case 11:
			goto tr612
		case 12:
			goto tr490
		case 13:
			goto tr586
		case 32:
			goto tr611
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st550
		}
		goto st42
	st550:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof550
		}
	st_case_550:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr611
		case 10:
			goto tr584
		case 11:
			goto tr612
		case 12:
			goto tr490
		case 13:
			goto tr586
		case 32:
			goto tr611
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st551
		}
		goto st42
	st551:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof551
		}
	st_case_551:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr611
		case 10:
			goto tr584
		case 11:
			goto tr612
		case 12:
			goto tr490
		case 13:
			goto tr586
		case 32:
			goto tr611
		case 34:
			goto tr128
		case 44:
			goto tr92
		case 61:
			goto tr129
		case 92:
			goto st94
		}
		goto st42
tr213:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st151
	st151:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof151
		}
	st_case_151:
//line plugins/parsers/influx/machine.go:21069
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr180
		case 10:
			goto st7
		case 11:
			goto tr181
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr180
		case 34:
			goto tr91
		case 44:
			goto tr182
		case 46:
			goto st152
		case 48:
			goto st576
		case 92:
			goto st157
		}
		if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st579
		}
		goto st55
tr214:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st152
	st152:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof152
		}
	st_case_152:
//line plugins/parsers/influx/machine.go:21109
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr180
		case 10:
			goto st7
		case 11:
			goto tr181
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr180
		case 34:
			goto tr91
		case 44:
			goto tr182
		case 92:
			goto st157
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st552
		}
		goto st55
	st552:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof552
		}
	st_case_552:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr837
		case 10:
			goto tr515
		case 11:
			goto tr838
		case 12:
			goto tr622
		case 13:
			goto tr517
		case 32:
			goto tr837
		case 34:
			goto tr91
		case 44:
			goto tr839
		case 69:
			goto st155
		case 92:
			goto st157
		case 101:
			goto st155
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st552
		}
		goto st55
tr838:
	( m.cs) = 553
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr870:
	( m.cs) = 553
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr874:
	( m.cs) = 553
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr878:
	( m.cs) = 553
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st553:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof553
		}
	st_case_553:
//line plugins/parsers/influx/machine.go:21264
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr841
		case 10:
			goto st317
		case 11:
			goto tr842
		case 12:
			goto tr482
		case 13:
			goto st104
		case 32:
			goto tr841
		case 34:
			goto tr124
		case 44:
			goto tr182
		case 45:
			goto tr843
		case 61:
			goto st55
		case 92:
			goto tr186
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr844
		}
		goto tr184
tr842:
	( m.cs) = 554
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
	st554:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof554
		}
	st_case_554:
//line plugins/parsers/influx/machine.go:21315
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr841
		case 10:
			goto st317
		case 11:
			goto tr842
		case 12:
			goto tr482
		case 13:
			goto st104
		case 32:
			goto tr841
		case 34:
			goto tr124
		case 44:
			goto tr182
		case 45:
			goto tr843
		case 61:
			goto tr189
		case 92:
			goto tr186
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr844
		}
		goto tr184
tr843:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st153
	st153:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof153
		}
	st_case_153:
//line plugins/parsers/influx/machine.go:21355
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr180
		case 10:
			goto st7
		case 11:
			goto tr188
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr180
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st555
		}
		goto st57
tr844:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st555
	st555:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof555
		}
	st_case_555:
//line plugins/parsers/influx/machine.go:21393
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr845
		case 10:
			goto tr659
		case 11:
			goto tr846
		case 12:
			goto tr490
		case 13:
			goto tr661
		case 32:
			goto tr845
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st557
		}
		goto st57
tr849:
	( m.cs) = 556
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto _again
tr846:
	( m.cs) = 556
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st556:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof556
		}
	st_case_556:
//line plugins/parsers/influx/machine.go:21465
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr848
		case 10:
			goto st317
		case 11:
			goto tr849
		case 12:
			goto tr495
		case 13:
			goto st104
		case 32:
			goto tr848
		case 34:
			goto tr124
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto tr186
		}
		goto tr184
tr186:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st154
	st154:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof154
		}
	st_case_154:
//line plugins/parsers/influx/machine.go:21500
		switch ( m.data)[( m.p)] {
		case 34:
			goto st57
		case 92:
			goto st57
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr8
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr8
		}
		goto st12
	st557:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof557
		}
	st_case_557:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr845
		case 10:
			goto tr659
		case 11:
			goto tr846
		case 12:
			goto tr490
		case 13:
			goto tr661
		case 32:
			goto tr845
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st558
		}
		goto st57
	st558:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof558
		}
	st_case_558:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr845
		case 10:
			goto tr659
		case 11:
			goto tr846
		case 12:
			goto tr490
		case 13:
			goto tr661
		case 32:
			goto tr845
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st559
		}
		goto st57
	st559:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof559
		}
	st_case_559:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr845
		case 10:
			goto tr659
		case 11:
			goto tr846
		case 12:
			goto tr490
		case 13:
			goto tr661
		case 32:
			goto tr845
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st560
		}
		goto st57
	st560:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof560
		}
	st_case_560:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr845
		case 10:
			goto tr659
		case 11:
			goto tr846
		case 12:
			goto tr490
		case 13:
			goto tr661
		case 32:
			goto tr845
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st561
		}
		goto st57
	st561:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof561
		}
	st_case_561:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr845
		case 10:
			goto tr659
		case 11:
			goto tr846
		case 12:
			goto tr490
		case 13:
			goto tr661
		case 32:
			goto tr845
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st562
		}
		goto st57
	st562:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof562
		}
	st_case_562:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr845
		case 10:
			goto tr659
		case 11:
			goto tr846
		case 12:
			goto tr490
		case 13:
			goto tr661
		case 32:
			goto tr845
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st563
		}
		goto st57
	st563:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof563
		}
	st_case_563:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr845
		case 10:
			goto tr659
		case 11:
			goto tr846
		case 12:
			goto tr490
		case 13:
			goto tr661
		case 32:
			goto tr845
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st564
		}
		goto st57
	st564:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof564
		}
	st_case_564:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr845
		case 10:
			goto tr659
		case 11:
			goto tr846
		case 12:
			goto tr490
		case 13:
			goto tr661
		case 32:
			goto tr845
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st565
		}
		goto st57
	st565:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof565
		}
	st_case_565:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr845
		case 10:
			goto tr659
		case 11:
			goto tr846
		case 12:
			goto tr490
		case 13:
			goto tr661
		case 32:
			goto tr845
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st566
		}
		goto st57
	st566:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof566
		}
	st_case_566:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr845
		case 10:
			goto tr659
		case 11:
			goto tr846
		case 12:
			goto tr490
		case 13:
			goto tr661
		case 32:
			goto tr845
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st567
		}
		goto st57
	st567:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof567
		}
	st_case_567:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr845
		case 10:
			goto tr659
		case 11:
			goto tr846
		case 12:
			goto tr490
		case 13:
			goto tr661
		case 32:
			goto tr845
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st568
		}
		goto st57
	st568:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof568
		}
	st_case_568:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr845
		case 10:
			goto tr659
		case 11:
			goto tr846
		case 12:
			goto tr490
		case 13:
			goto tr661
		case 32:
			goto tr845
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st569
		}
		goto st57
	st569:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof569
		}
	st_case_569:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr845
		case 10:
			goto tr659
		case 11:
			goto tr846
		case 12:
			goto tr490
		case 13:
			goto tr661
		case 32:
			goto tr845
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st570
		}
		goto st57
	st570:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof570
		}
	st_case_570:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr845
		case 10:
			goto tr659
		case 11:
			goto tr846
		case 12:
			goto tr490
		case 13:
			goto tr661
		case 32:
			goto tr845
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st571
		}
		goto st57
	st571:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof571
		}
	st_case_571:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr845
		case 10:
			goto tr659
		case 11:
			goto tr846
		case 12:
			goto tr490
		case 13:
			goto tr661
		case 32:
			goto tr845
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st572
		}
		goto st57
	st572:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof572
		}
	st_case_572:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr845
		case 10:
			goto tr659
		case 11:
			goto tr846
		case 12:
			goto tr490
		case 13:
			goto tr661
		case 32:
			goto tr845
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st573
		}
		goto st57
	st573:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof573
		}
	st_case_573:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr845
		case 10:
			goto tr659
		case 11:
			goto tr846
		case 12:
			goto tr490
		case 13:
			goto tr661
		case 32:
			goto tr845
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st574
		}
		goto st57
	st574:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof574
		}
	st_case_574:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr845
		case 10:
			goto tr659
		case 11:
			goto tr846
		case 12:
			goto tr490
		case 13:
			goto tr661
		case 32:
			goto tr845
		case 34:
			goto tr128
		case 44:
			goto tr182
		case 61:
			goto tr189
		case 92:
			goto st154
		}
		goto st57
	st155:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof155
		}
	st_case_155:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr180
		case 10:
			goto st7
		case 11:
			goto tr181
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr180
		case 34:
			goto tr317
		case 44:
			goto tr182
		case 92:
			goto st157
		}
		switch {
		case ( m.data)[( m.p)] > 45:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st575
			}
		case ( m.data)[( m.p)] >= 43:
			goto st156
		}
		goto st55
	st156:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof156
		}
	st_case_156:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr180
		case 10:
			goto st7
		case 11:
			goto tr181
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr180
		case 34:
			goto tr91
		case 44:
			goto tr182
		case 92:
			goto st157
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st575
		}
		goto st55
	st575:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof575
		}
	st_case_575:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr837
		case 10:
			goto tr515
		case 11:
			goto tr838
		case 12:
			goto tr622
		case 13:
			goto tr517
		case 32:
			goto tr837
		case 34:
			goto tr91
		case 44:
			goto tr839
		case 92:
			goto st157
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st575
		}
		goto st55
tr340:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st157
	st157:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof157
		}
	st_case_157:
//line plugins/parsers/influx/machine.go:22174
		switch ( m.data)[( m.p)] {
		case 34:
			goto st55
		case 92:
			goto st55
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr8
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr8
		}
		goto st1
	st576:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof576
		}
	st_case_576:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr837
		case 10:
			goto tr515
		case 11:
			goto tr838
		case 12:
			goto tr622
		case 13:
			goto tr517
		case 32:
			goto tr837
		case 34:
			goto tr91
		case 44:
			goto tr839
		case 46:
			goto st552
		case 69:
			goto st155
		case 92:
			goto st157
		case 101:
			goto st155
		case 105:
			goto st578
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st577
		}
		goto st55
	st577:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof577
		}
	st_case_577:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr837
		case 10:
			goto tr515
		case 11:
			goto tr838
		case 12:
			goto tr622
		case 13:
			goto tr517
		case 32:
			goto tr837
		case 34:
			goto tr91
		case 44:
			goto tr839
		case 46:
			goto st552
		case 69:
			goto st155
		case 92:
			goto st157
		case 101:
			goto st155
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st577
		}
		goto st55
	st578:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof578
		}
	st_case_578:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr869
		case 10:
			goto tr722
		case 11:
			goto tr870
		case 12:
			goto tr804
		case 13:
			goto tr725
		case 32:
			goto tr869
		case 34:
			goto tr91
		case 44:
			goto tr871
		case 92:
			goto st157
		}
		goto st55
	st579:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof579
		}
	st_case_579:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr837
		case 10:
			goto tr515
		case 11:
			goto tr838
		case 12:
			goto tr622
		case 13:
			goto tr517
		case 32:
			goto tr837
		case 34:
			goto tr91
		case 44:
			goto tr839
		case 46:
			goto st552
		case 69:
			goto st155
		case 92:
			goto st157
		case 101:
			goto st155
		case 105:
			goto st578
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st579
		}
		goto st55
tr215:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st580
	st580:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof580
		}
	st_case_580:
//line plugins/parsers/influx/machine.go:22336
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr837
		case 10:
			goto tr515
		case 11:
			goto tr838
		case 12:
			goto tr622
		case 13:
			goto tr517
		case 32:
			goto tr837
		case 34:
			goto tr91
		case 44:
			goto tr839
		case 46:
			goto st552
		case 69:
			goto st155
		case 92:
			goto st157
		case 101:
			goto st155
		case 105:
			goto st578
		case 117:
			goto st581
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st577
		}
		goto st55
	st581:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof581
		}
	st_case_581:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr873
		case 10:
			goto tr729
		case 11:
			goto tr874
		case 12:
			goto tr809
		case 13:
			goto tr732
		case 32:
			goto tr873
		case 34:
			goto tr91
		case 44:
			goto tr875
		case 92:
			goto st157
		}
		goto st55
tr216:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st582
	st582:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof582
		}
	st_case_582:
//line plugins/parsers/influx/machine.go:22408
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr837
		case 10:
			goto tr515
		case 11:
			goto tr838
		case 12:
			goto tr622
		case 13:
			goto tr517
		case 32:
			goto tr837
		case 34:
			goto tr91
		case 44:
			goto tr839
		case 46:
			goto st552
		case 69:
			goto st155
		case 92:
			goto st157
		case 101:
			goto st155
		case 105:
			goto st578
		case 117:
			goto st581
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st582
		}
		goto st55
tr217:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st583
	st583:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof583
		}
	st_case_583:
//line plugins/parsers/influx/machine.go:22454
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr877
		case 10:
			goto tr736
		case 11:
			goto tr878
		case 12:
			goto tr814
		case 13:
			goto tr739
		case 32:
			goto tr877
		case 34:
			goto tr91
		case 44:
			goto tr879
		case 65:
			goto st158
		case 92:
			goto st157
		case 97:
			goto st161
		}
		goto st55
	st158:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof158
		}
	st_case_158:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr180
		case 10:
			goto st7
		case 11:
			goto tr181
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr180
		case 34:
			goto tr91
		case 44:
			goto tr182
		case 76:
			goto st159
		case 92:
			goto st157
		}
		goto st55
	st159:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof159
		}
	st_case_159:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr180
		case 10:
			goto st7
		case 11:
			goto tr181
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr180
		case 34:
			goto tr91
		case 44:
			goto tr182
		case 83:
			goto st160
		case 92:
			goto st157
		}
		goto st55
	st160:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof160
		}
	st_case_160:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr180
		case 10:
			goto st7
		case 11:
			goto tr181
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr180
		case 34:
			goto tr91
		case 44:
			goto tr182
		case 69:
			goto st584
		case 92:
			goto st157
		}
		goto st55
	st584:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof584
		}
	st_case_584:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr877
		case 10:
			goto tr736
		case 11:
			goto tr878
		case 12:
			goto tr814
		case 13:
			goto tr739
		case 32:
			goto tr877
		case 34:
			goto tr91
		case 44:
			goto tr879
		case 92:
			goto st157
		}
		goto st55
	st161:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof161
		}
	st_case_161:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr180
		case 10:
			goto st7
		case 11:
			goto tr181
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr180
		case 34:
			goto tr91
		case 44:
			goto tr182
		case 92:
			goto st157
		case 108:
			goto st162
		}
		goto st55
	st162:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof162
		}
	st_case_162:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr180
		case 10:
			goto st7
		case 11:
			goto tr181
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr180
		case 34:
			goto tr91
		case 44:
			goto tr182
		case 92:
			goto st157
		case 115:
			goto st163
		}
		goto st55
	st163:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof163
		}
	st_case_163:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr180
		case 10:
			goto st7
		case 11:
			goto tr181
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr180
		case 34:
			goto tr91
		case 44:
			goto tr182
		case 92:
			goto st157
		case 101:
			goto st584
		}
		goto st55
tr218:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st585
	st585:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof585
		}
	st_case_585:
//line plugins/parsers/influx/machine.go:22685
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr877
		case 10:
			goto tr736
		case 11:
			goto tr878
		case 12:
			goto tr814
		case 13:
			goto tr739
		case 32:
			goto tr877
		case 34:
			goto tr91
		case 44:
			goto tr879
		case 82:
			goto st164
		case 92:
			goto st157
		case 114:
			goto st165
		}
		goto st55
	st164:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof164
		}
	st_case_164:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr180
		case 10:
			goto st7
		case 11:
			goto tr181
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr180
		case 34:
			goto tr91
		case 44:
			goto tr182
		case 85:
			goto st160
		case 92:
			goto st157
		}
		goto st55
	st165:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof165
		}
	st_case_165:
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr180
		case 10:
			goto st7
		case 11:
			goto tr181
		case 12:
			goto tr1
		case 13:
			goto st8
		case 32:
			goto tr180
		case 34:
			goto tr91
		case 44:
			goto tr182
		case 92:
			goto st157
		case 117:
			goto st163
		}
		goto st55
tr219:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st586
	st586:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof586
		}
	st_case_586:
//line plugins/parsers/influx/machine.go:22778
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr877
		case 10:
			goto tr736
		case 11:
			goto tr878
		case 12:
			goto tr814
		case 13:
			goto tr739
		case 32:
			goto tr877
		case 34:
			goto tr91
		case 44:
			goto tr879
		case 92:
			goto st157
		case 97:
			goto st161
		}
		goto st55
tr220:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st587
	st587:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof587
		}
	st_case_587:
//line plugins/parsers/influx/machine.go:22813
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr877
		case 10:
			goto tr736
		case 11:
			goto tr878
		case 12:
			goto tr814
		case 13:
			goto tr739
		case 32:
			goto tr877
		case 34:
			goto tr91
		case 44:
			goto tr879
		case 92:
			goto st157
		case 114:
			goto st165
		}
		goto st55
	st166:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof166
		}
	st_case_166:
		switch ( m.data)[( m.p)] {
		case 9:
			goto st166
		case 10:
			goto st7
		case 11:
			goto tr339
		case 12:
			goto st9
		case 13:
			goto st8
		case 32:
			goto st166
		case 34:
			goto tr118
		case 35:
			goto st6
		case 44:
			goto st6
		case 92:
			goto tr340
		}
		goto tr337
tr339:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st167
	st167:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof167
		}
	st_case_167:
//line plugins/parsers/influx/machine.go:22876
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr341
		case 10:
			goto st7
		case 11:
			goto tr342
		case 12:
			goto tr38
		case 13:
			goto st8
		case 32:
			goto tr341
		case 34:
			goto tr85
		case 35:
			goto st55
		case 44:
			goto tr182
		case 92:
			goto tr340
		}
		goto tr337
tr341:
	( m.cs) = 168
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st168:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof168
		}
	st_case_168:
//line plugins/parsers/influx/machine.go:22918
		switch ( m.data)[( m.p)] {
		case 9:
			goto st168
		case 10:
			goto st7
		case 11:
			goto tr344
		case 12:
			goto st11
		case 13:
			goto st8
		case 32:
			goto st168
		case 34:
			goto tr124
		case 35:
			goto tr160
		case 44:
			goto st6
		case 61:
			goto tr337
		case 92:
			goto tr186
		}
		goto tr184
tr344:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st169
tr345:
	( m.cs) = 169
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st169:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof169
		}
	st_case_169:
//line plugins/parsers/influx/machine.go:22972
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr341
		case 10:
			goto st7
		case 11:
			goto tr345
		case 12:
			goto tr38
		case 13:
			goto st8
		case 32:
			goto tr341
		case 34:
			goto tr124
		case 44:
			goto tr182
		case 61:
			goto tr346
		case 92:
			goto tr186
		}
		goto tr184
tr342:
	( m.cs) = 170
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st170:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof170
		}
	st_case_170:
//line plugins/parsers/influx/machine.go:23018
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr341
		case 10:
			goto st7
		case 11:
			goto tr345
		case 12:
			goto tr38
		case 13:
			goto st8
		case 32:
			goto tr341
		case 34:
			goto tr124
		case 44:
			goto tr182
		case 61:
			goto tr337
		case 92:
			goto tr186
		}
		goto tr184
tr522:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st171
	st171:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof171
		}
	st_case_171:
//line plugins/parsers/influx/machine.go:23053
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr105
		case 13:
			goto st8
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st588
		}
		goto st6
tr523:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st588
	st588:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof588
		}
	st_case_588:
//line plugins/parsers/influx/machine.go:23081
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr659
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st589
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr658
		}
		goto st6
	st589:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof589
		}
	st_case_589:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr659
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st590
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr658
		}
		goto st6
	st590:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof590
		}
	st_case_590:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr659
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st591
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr658
		}
		goto st6
	st591:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof591
		}
	st_case_591:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr659
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st592
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr658
		}
		goto st6
	st592:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof592
		}
	st_case_592:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr659
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st593
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr658
		}
		goto st6
	st593:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof593
		}
	st_case_593:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr659
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st594
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr658
		}
		goto st6
	st594:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof594
		}
	st_case_594:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr659
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st595
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr658
		}
		goto st6
	st595:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof595
		}
	st_case_595:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr659
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st596
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr658
		}
		goto st6
	st596:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof596
		}
	st_case_596:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr659
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st597
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr658
		}
		goto st6
	st597:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof597
		}
	st_case_597:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr659
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st598
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr658
		}
		goto st6
	st598:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof598
		}
	st_case_598:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr659
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st599
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr658
		}
		goto st6
	st599:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof599
		}
	st_case_599:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr659
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st600
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr658
		}
		goto st6
	st600:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof600
		}
	st_case_600:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr659
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st601
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr658
		}
		goto st6
	st601:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof601
		}
	st_case_601:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr659
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st602
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr658
		}
		goto st6
	st602:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof602
		}
	st_case_602:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr659
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st603
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr658
		}
		goto st6
	st603:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof603
		}
	st_case_603:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr659
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st604
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr658
		}
		goto st6
	st604:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof604
		}
	st_case_604:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr659
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st605
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr658
		}
		goto st6
	st605:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof605
		}
	st_case_605:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr659
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st606
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr658
		}
		goto st6
	st606:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof606
		}
	st_case_606:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr659
		case 12:
			goto tr450
		case 13:
			goto tr661
		case 32:
			goto tr658
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr658
		}
		goto st6
tr903:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st172
tr518:
	( m.cs) = 172
//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr910:
	( m.cs) = 172
//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr913:
	( m.cs) = 172
//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr917:
	( m.cs) = 172
//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st172:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof172
		}
	st_case_172:
//line plugins/parsers/influx/machine.go:23667
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr97
		case 44:
			goto st6
		case 61:
			goto st6
		case 92:
			goto tr349
		}
		goto tr348
tr348:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st173
	st173:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof173
		}
	st_case_173:
//line plugins/parsers/influx/machine.go:23700
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr351
		case 92:
			goto st185
		}
		goto st173
tr351:
//line plugins/parsers/influx/machine.go.rl:99

	key = m.text()

	goto st174
	st174:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof174
		}
	st_case_174:
//line plugins/parsers/influx/machine.go:23733
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr353
		case 45:
			goto tr167
		case 46:
			goto tr168
		case 48:
			goto tr169
		case 70:
			goto tr171
		case 84:
			goto tr172
		case 92:
			goto st76
		case 102:
			goto tr173
		case 116:
			goto tr174
		}
		if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr170
		}
		goto st6
tr353:
	( m.cs) = 607
//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st607:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof607
		}
	st_case_607:
//line plugins/parsers/influx/machine.go:23782
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr650
		case 12:
			goto st261
		case 13:
			goto tr652
		case 32:
			goto tr902
		case 34:
			goto tr26
		case 44:
			goto tr903
		case 92:
			goto tr27
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr902
		}
		goto tr23
tr169:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st608
	st608:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof608
		}
	st_case_608:
//line plugins/parsers/influx/machine.go:23814
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr515
		case 12:
			goto tr516
		case 13:
			goto tr517
		case 32:
			goto tr514
		case 34:
			goto tr31
		case 44:
			goto tr518
		case 46:
			goto st315
		case 69:
			goto st175
		case 92:
			goto st76
		case 101:
			goto st175
		case 105:
			goto st613
		case 117:
			goto st614
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st609
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr514
		}
		goto st6
	st609:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof609
		}
	st_case_609:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr515
		case 12:
			goto tr516
		case 13:
			goto tr517
		case 32:
			goto tr514
		case 34:
			goto tr31
		case 44:
			goto tr518
		case 46:
			goto st315
		case 69:
			goto st175
		case 92:
			goto st76
		case 101:
			goto st175
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st609
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr514
		}
		goto st6
	st175:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof175
		}
	st_case_175:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr354
		case 43:
			goto st176
		case 45:
			goto st176
		case 92:
			goto st76
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st612
		}
		goto st6
tr354:
	( m.cs) = 610
//line plugins/parsers/influx/machine.go.rl:139

	err = m.handler.AddString(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st610:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof610
		}
	st_case_610:
//line plugins/parsers/influx/machine.go:23929
		switch ( m.data)[( m.p)] {
		case 10:
			goto st262
		case 13:
			goto st34
		case 32:
			goto st261
		case 44:
			goto st37
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st611
			}
		case ( m.data)[( m.p)] >= 9:
			goto st261
		}
		goto tr105
	st611:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof611
		}
	st_case_611:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 13:
			goto tr717
		case 32:
			goto tr516
		case 44:
			goto tr907
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st611
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr516
		}
		goto tr105
	st176:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof176
		}
	st_case_176:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st612
		}
		goto st6
	st612:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof612
		}
	st_case_612:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr515
		case 12:
			goto tr516
		case 13:
			goto tr517
		case 32:
			goto tr514
		case 34:
			goto tr31
		case 44:
			goto tr518
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st612
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr514
		}
		goto st6
	st613:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof613
		}
	st_case_613:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr722
		case 12:
			goto tr909
		case 13:
			goto tr725
		case 32:
			goto tr908
		case 34:
			goto tr31
		case 44:
			goto tr910
		case 92:
			goto st76
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr908
		}
		goto st6
	st614:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof614
		}
	st_case_614:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr729
		case 12:
			goto tr912
		case 13:
			goto tr732
		case 32:
			goto tr911
		case 34:
			goto tr31
		case 44:
			goto tr913
		case 92:
			goto st76
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr911
		}
		goto st6
tr170:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st615
	st615:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof615
		}
	st_case_615:
//line plugins/parsers/influx/machine.go:24085
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr515
		case 12:
			goto tr516
		case 13:
			goto tr517
		case 32:
			goto tr514
		case 34:
			goto tr31
		case 44:
			goto tr518
		case 46:
			goto st315
		case 69:
			goto st175
		case 92:
			goto st76
		case 101:
			goto st175
		case 105:
			goto st613
		case 117:
			goto st614
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st615
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr514
		}
		goto st6
tr171:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st616
	st616:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof616
		}
	st_case_616:
//line plugins/parsers/influx/machine.go:24132
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr736
		case 12:
			goto tr916
		case 13:
			goto tr739
		case 32:
			goto tr915
		case 34:
			goto tr31
		case 44:
			goto tr917
		case 65:
			goto st177
		case 92:
			goto st76
		case 97:
			goto st180
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr915
		}
		goto st6
	st177:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof177
		}
	st_case_177:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 76:
			goto st178
		case 92:
			goto st76
		}
		goto st6
	st178:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof178
		}
	st_case_178:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 83:
			goto st179
		case 92:
			goto st76
		}
		goto st6
	st179:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof179
		}
	st_case_179:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 69:
			goto st617
		case 92:
			goto st76
		}
		goto st6
	st617:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof617
		}
	st_case_617:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr736
		case 12:
			goto tr916
		case 13:
			goto tr739
		case 32:
			goto tr915
		case 34:
			goto tr31
		case 44:
			goto tr917
		case 92:
			goto st76
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr915
		}
		goto st6
	st180:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof180
		}
	st_case_180:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 92:
			goto st76
		case 108:
			goto st181
		}
		goto st6
	st181:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof181
		}
	st_case_181:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 92:
			goto st76
		case 115:
			goto st182
		}
		goto st6
	st182:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof182
		}
	st_case_182:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 92:
			goto st76
		case 101:
			goto st617
		}
		goto st6
tr172:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st618
	st618:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof618
		}
	st_case_618:
//line plugins/parsers/influx/machine.go:24313
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr736
		case 12:
			goto tr916
		case 13:
			goto tr739
		case 32:
			goto tr915
		case 34:
			goto tr31
		case 44:
			goto tr917
		case 82:
			goto st183
		case 92:
			goto st76
		case 114:
			goto st184
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr915
		}
		goto st6
	st183:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof183
		}
	st_case_183:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 85:
			goto st179
		case 92:
			goto st76
		}
		goto st6
	st184:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof184
		}
	st_case_184:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 92:
			goto st76
		case 117:
			goto st182
		}
		goto st6
tr173:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st619
	st619:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof619
		}
	st_case_619:
//line plugins/parsers/influx/machine.go:24389
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr736
		case 12:
			goto tr916
		case 13:
			goto tr739
		case 32:
			goto tr915
		case 34:
			goto tr31
		case 44:
			goto tr917
		case 92:
			goto st76
		case 97:
			goto st180
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr915
		}
		goto st6
tr174:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st620
	st620:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof620
		}
	st_case_620:
//line plugins/parsers/influx/machine.go:24423
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr736
		case 12:
			goto tr916
		case 13:
			goto tr739
		case 32:
			goto tr915
		case 34:
			goto tr31
		case 44:
			goto tr917
		case 92:
			goto st76
		case 114:
			goto st184
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr915
		}
		goto st6
tr349:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st185
	st185:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof185
		}
	st_case_185:
//line plugins/parsers/influx/machine.go:24457
		switch ( m.data)[( m.p)] {
		case 34:
			goto st173
		case 92:
			goto st173
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr8
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr8
		}
		goto st3
	st621:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof621
		}
	st_case_621:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr515
		case 12:
			goto tr516
		case 13:
			goto tr517
		case 32:
			goto tr514
		case 34:
			goto tr31
		case 44:
			goto tr518
		case 46:
			goto st315
		case 69:
			goto st175
		case 92:
			goto st76
		case 101:
			goto st175
		case 105:
			goto st613
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st609
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr514
		}
		goto st6
	st622:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof622
		}
	st_case_622:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr515
		case 12:
			goto tr516
		case 13:
			goto tr517
		case 32:
			goto tr514
		case 34:
			goto tr31
		case 44:
			goto tr518
		case 46:
			goto st315
		case 69:
			goto st175
		case 92:
			goto st76
		case 101:
			goto st175
		case 105:
			goto st613
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st622
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr514
		}
		goto st6
tr162:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st186
	st186:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof186
		}
	st_case_186:
//line plugins/parsers/influx/machine.go:24560
		switch ( m.data)[( m.p)] {
		case 9:
			goto st50
		case 10:
			goto st7
		case 11:
			goto tr162
		case 12:
			goto st2
		case 13:
			goto st8
		case 32:
			goto st50
		case 34:
			goto tr97
		case 44:
			goto st6
		case 61:
			goto tr165
		case 92:
			goto tr163
		}
		goto tr160
tr140:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st187
	st187:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof187
		}
	st_case_187:
//line plugins/parsers/influx/machine.go:24595
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr61
		case 13:
			goto tr47
		case 32:
			goto tr60
		case 44:
			goto tr62
		case 46:
			goto st188
		case 48:
			goto st624
		case 61:
			goto tr47
		case 92:
			goto st23
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st627
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr60
		}
		goto st17
tr141:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st188
	st188:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof188
		}
	st_case_188:
//line plugins/parsers/influx/machine.go:24636
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr61
		case 13:
			goto tr47
		case 32:
			goto tr60
		case 44:
			goto tr62
		case 61:
			goto tr47
		case 92:
			goto st23
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st623
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr60
		}
		goto st17
	st623:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof623
		}
	st_case_623:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 11:
			goto tr716
		case 13:
			goto tr717
		case 32:
			goto tr712
		case 44:
			goto tr718
		case 61:
			goto tr132
		case 69:
			goto st189
		case 92:
			goto st23
		case 101:
			goto st189
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st623
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr712
		}
		goto st17
	st189:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof189
		}
	st_case_189:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr61
		case 13:
			goto tr47
		case 32:
			goto tr60
		case 34:
			goto st190
		case 44:
			goto tr62
		case 61:
			goto tr47
		case 92:
			goto st23
		}
		switch {
		case ( m.data)[( m.p)] < 43:
			if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
				goto tr60
			}
		case ( m.data)[( m.p)] > 45:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st469
			}
		default:
			goto st190
		}
		goto st17
	st190:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof190
		}
	st_case_190:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr61
		case 13:
			goto tr47
		case 32:
			goto tr60
		case 44:
			goto tr62
		case 61:
			goto tr47
		case 92:
			goto st23
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st469
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr60
		}
		goto st17
	st624:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof624
		}
	st_case_624:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 11:
			goto tr716
		case 13:
			goto tr717
		case 32:
			goto tr712
		case 44:
			goto tr718
		case 46:
			goto st623
		case 61:
			goto tr132
		case 69:
			goto st189
		case 92:
			goto st23
		case 101:
			goto st189
		case 105:
			goto st626
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st625
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr712
		}
		goto st17
	st625:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof625
		}
	st_case_625:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 11:
			goto tr716
		case 13:
			goto tr717
		case 32:
			goto tr712
		case 44:
			goto tr718
		case 46:
			goto st623
		case 61:
			goto tr132
		case 69:
			goto st189
		case 92:
			goto st23
		case 101:
			goto st189
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st625
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr712
		}
		goto st17
	st626:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof626
		}
	st_case_626:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr925
		case 11:
			goto tr926
		case 13:
			goto tr927
		case 32:
			goto tr724
		case 44:
			goto tr928
		case 61:
			goto tr132
		case 92:
			goto st23
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr724
		}
		goto st17
	st627:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof627
		}
	st_case_627:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 11:
			goto tr716
		case 13:
			goto tr717
		case 32:
			goto tr712
		case 44:
			goto tr718
		case 46:
			goto st623
		case 61:
			goto tr132
		case 69:
			goto st189
		case 92:
			goto st23
		case 101:
			goto st189
		case 105:
			goto st626
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st627
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr712
		}
		goto st17
tr142:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st628
	st628:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof628
		}
	st_case_628:
//line plugins/parsers/influx/machine.go:24910
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 11:
			goto tr716
		case 13:
			goto tr717
		case 32:
			goto tr712
		case 44:
			goto tr718
		case 46:
			goto st623
		case 61:
			goto tr132
		case 69:
			goto st189
		case 92:
			goto st23
		case 101:
			goto st189
		case 105:
			goto st626
		case 117:
			goto st629
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st625
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr712
		}
		goto st17
	st629:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof629
		}
	st_case_629:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr930
		case 11:
			goto tr931
		case 13:
			goto tr932
		case 32:
			goto tr731
		case 44:
			goto tr933
		case 61:
			goto tr132
		case 92:
			goto st23
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr731
		}
		goto st17
tr143:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st630
	st630:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof630
		}
	st_case_630:
//line plugins/parsers/influx/machine.go:24982
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 11:
			goto tr716
		case 13:
			goto tr717
		case 32:
			goto tr712
		case 44:
			goto tr718
		case 46:
			goto st623
		case 61:
			goto tr132
		case 69:
			goto st189
		case 92:
			goto st23
		case 101:
			goto st189
		case 105:
			goto st626
		case 117:
			goto st629
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st630
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr712
		}
		goto st17
tr144:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st631
	st631:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof631
		}
	st_case_631:
//line plugins/parsers/influx/machine.go:25029
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr935
		case 11:
			goto tr936
		case 13:
			goto tr937
		case 32:
			goto tr738
		case 44:
			goto tr938
		case 61:
			goto tr132
		case 65:
			goto st191
		case 92:
			goto st23
		case 97:
			goto st194
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr738
		}
		goto st17
	st191:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof191
		}
	st_case_191:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr61
		case 13:
			goto tr47
		case 32:
			goto tr60
		case 44:
			goto tr62
		case 61:
			goto tr47
		case 76:
			goto st192
		case 92:
			goto st23
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr60
		}
		goto st17
	st192:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof192
		}
	st_case_192:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr61
		case 13:
			goto tr47
		case 32:
			goto tr60
		case 44:
			goto tr62
		case 61:
			goto tr47
		case 83:
			goto st193
		case 92:
			goto st23
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr60
		}
		goto st17
	st193:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof193
		}
	st_case_193:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr61
		case 13:
			goto tr47
		case 32:
			goto tr60
		case 44:
			goto tr62
		case 61:
			goto tr47
		case 69:
			goto st632
		case 92:
			goto st23
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr60
		}
		goto st17
	st632:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof632
		}
	st_case_632:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr935
		case 11:
			goto tr936
		case 13:
			goto tr937
		case 32:
			goto tr738
		case 44:
			goto tr938
		case 61:
			goto tr132
		case 92:
			goto st23
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr738
		}
		goto st17
	st194:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof194
		}
	st_case_194:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr61
		case 13:
			goto tr47
		case 32:
			goto tr60
		case 44:
			goto tr62
		case 61:
			goto tr47
		case 92:
			goto st23
		case 108:
			goto st195
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr60
		}
		goto st17
	st195:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof195
		}
	st_case_195:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr61
		case 13:
			goto tr47
		case 32:
			goto tr60
		case 44:
			goto tr62
		case 61:
			goto tr47
		case 92:
			goto st23
		case 115:
			goto st196
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr60
		}
		goto st17
	st196:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof196
		}
	st_case_196:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr61
		case 13:
			goto tr47
		case 32:
			goto tr60
		case 44:
			goto tr62
		case 61:
			goto tr47
		case 92:
			goto st23
		case 101:
			goto st632
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr60
		}
		goto st17
tr145:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st633
	st633:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof633
		}
	st_case_633:
//line plugins/parsers/influx/machine.go:25252
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr935
		case 11:
			goto tr936
		case 13:
			goto tr937
		case 32:
			goto tr738
		case 44:
			goto tr938
		case 61:
			goto tr132
		case 82:
			goto st197
		case 92:
			goto st23
		case 114:
			goto st198
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr738
		}
		goto st17
	st197:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof197
		}
	st_case_197:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr61
		case 13:
			goto tr47
		case 32:
			goto tr60
		case 44:
			goto tr62
		case 61:
			goto tr47
		case 85:
			goto st193
		case 92:
			goto st23
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr60
		}
		goto st17
	st198:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof198
		}
	st_case_198:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr61
		case 13:
			goto tr47
		case 32:
			goto tr60
		case 44:
			goto tr62
		case 61:
			goto tr47
		case 92:
			goto st23
		case 117:
			goto st196
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr60
		}
		goto st17
tr146:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st634
	st634:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof634
		}
	st_case_634:
//line plugins/parsers/influx/machine.go:25342
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr935
		case 11:
			goto tr936
		case 13:
			goto tr937
		case 32:
			goto tr738
		case 44:
			goto tr938
		case 61:
			goto tr132
		case 92:
			goto st23
		case 97:
			goto st194
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr738
		}
		goto st17
tr147:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st635
	st635:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof635
		}
	st_case_635:
//line plugins/parsers/influx/machine.go:25376
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr935
		case 11:
			goto tr936
		case 13:
			goto tr937
		case 32:
			goto tr738
		case 44:
			goto tr938
		case 61:
			goto tr132
		case 92:
			goto st23
		case 114:
			goto st198
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr738
		}
		goto st17
tr123:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st199
tr373:
	( m.cs) = 199
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st199:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof199
		}
	st_case_199:
//line plugins/parsers/influx/machine.go:25427
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr119
		case 10:
			goto st7
		case 11:
			goto tr373
		case 12:
			goto tr38
		case 13:
			goto st8
		case 32:
			goto tr119
		case 34:
			goto tr124
		case 44:
			goto tr92
		case 61:
			goto tr374
		case 92:
			goto tr125
		}
		goto tr121
tr120:
	( m.cs) = 200
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st200:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof200
		}
	st_case_200:
//line plugins/parsers/influx/machine.go:25473
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr119
		case 10:
			goto st7
		case 11:
			goto tr373
		case 12:
			goto tr38
		case 13:
			goto st8
		case 32:
			goto tr119
		case 34:
			goto tr124
		case 44:
			goto tr92
		case 61:
			goto tr82
		case 92:
			goto tr125
		}
		goto tr121
tr480:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st201
	st201:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof201
		}
	st_case_201:
//line plugins/parsers/influx/machine.go:25508
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr105
		case 13:
			goto st8
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st636
		}
		goto st6
tr481:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st636
	st636:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof636
		}
	st_case_636:
//line plugins/parsers/influx/machine.go:25536
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr584
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st637
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr583
		}
		goto st6
	st637:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof637
		}
	st_case_637:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr584
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st638
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr583
		}
		goto st6
	st638:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof638
		}
	st_case_638:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr584
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st639
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr583
		}
		goto st6
	st639:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof639
		}
	st_case_639:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr584
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st640
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr583
		}
		goto st6
	st640:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof640
		}
	st_case_640:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr584
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st641
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr583
		}
		goto st6
	st641:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof641
		}
	st_case_641:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr584
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st642
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr583
		}
		goto st6
	st642:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof642
		}
	st_case_642:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr584
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st643
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr583
		}
		goto st6
	st643:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof643
		}
	st_case_643:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr584
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st644
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr583
		}
		goto st6
	st644:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof644
		}
	st_case_644:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr584
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st645
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr583
		}
		goto st6
	st645:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof645
		}
	st_case_645:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr584
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st646
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr583
		}
		goto st6
	st646:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof646
		}
	st_case_646:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr584
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st647
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr583
		}
		goto st6
	st647:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof647
		}
	st_case_647:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr584
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st648
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr583
		}
		goto st6
	st648:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof648
		}
	st_case_648:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr584
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st649
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr583
		}
		goto st6
	st649:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof649
		}
	st_case_649:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr584
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st650
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr583
		}
		goto st6
	st650:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof650
		}
	st_case_650:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr584
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st651
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr583
		}
		goto st6
	st651:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof651
		}
	st_case_651:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr584
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st652
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr583
		}
		goto st6
	st652:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof652
		}
	st_case_652:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr584
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st653
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr583
		}
		goto st6
	st653:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof653
		}
	st_case_653:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr584
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st654
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr583
		}
		goto st6
	st654:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof654
		}
	st_case_654:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr584
		case 12:
			goto tr450
		case 13:
			goto tr586
		case 32:
			goto tr583
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr583
		}
		goto st6
tr477:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st202
tr962:
	( m.cs) = 202
//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr967:
	( m.cs) = 202
//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr970:
	( m.cs) = 202
//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr973:
	( m.cs) = 202
//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st202:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof202
		}
	st_case_202:
//line plugins/parsers/influx/machine.go:26122
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr377
		case 44:
			goto st6
		case 61:
			goto st6
		case 92:
			goto tr378
		}
		goto tr376
tr376:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st203
	st203:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof203
		}
	st_case_203:
//line plugins/parsers/influx/machine.go:26155
		switch ( m.data)[( m.p)] {
		case 9:
			goto st6
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 32:
			goto st6
		case 34:
			goto tr100
		case 44:
			goto st6
		case 61:
			goto tr380
		case 92:
			goto st217
		}
		goto st203
tr380:
//line plugins/parsers/influx/machine.go.rl:99

	key = m.text()

	goto st204
	st204:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof204
		}
	st_case_204:
//line plugins/parsers/influx/machine.go:26188
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr353
		case 45:
			goto tr108
		case 46:
			goto tr109
		case 48:
			goto tr110
		case 70:
			goto tr112
		case 84:
			goto tr113
		case 92:
			goto st76
		case 102:
			goto tr114
		case 116:
			goto tr115
		}
		if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto tr111
		}
		goto st6
tr108:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st205
	st205:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof205
		}
	st_case_205:
//line plugins/parsers/influx/machine.go:26230
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 46:
			goto st206
		case 48:
			goto st657
		case 92:
			goto st76
		}
		if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st660
		}
		goto st6
tr109:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st206
	st206:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof206
		}
	st_case_206:
//line plugins/parsers/influx/machine.go:26262
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st655
		}
		goto st6
	st655:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof655
		}
	st_case_655:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr620
		case 12:
			goto tr516
		case 13:
			goto tr623
		case 32:
			goto tr961
		case 34:
			goto tr31
		case 44:
			goto tr962
		case 69:
			goto st207
		case 92:
			goto st76
		case 101:
			goto st207
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st655
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr961
		}
		goto st6
	st207:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof207
		}
	st_case_207:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr354
		case 43:
			goto st208
		case 45:
			goto st208
		case 92:
			goto st76
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st656
		}
		goto st6
	st208:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof208
		}
	st_case_208:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 92:
			goto st76
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st656
		}
		goto st6
	st656:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof656
		}
	st_case_656:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr620
		case 12:
			goto tr516
		case 13:
			goto tr623
		case 32:
			goto tr961
		case 34:
			goto tr31
		case 44:
			goto tr962
		case 92:
			goto st76
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st656
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr961
		}
		goto st6
	st657:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof657
		}
	st_case_657:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr620
		case 12:
			goto tr516
		case 13:
			goto tr623
		case 32:
			goto tr961
		case 34:
			goto tr31
		case 44:
			goto tr962
		case 46:
			goto st655
		case 69:
			goto st207
		case 92:
			goto st76
		case 101:
			goto st207
		case 105:
			goto st659
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st658
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr961
		}
		goto st6
	st658:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof658
		}
	st_case_658:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr620
		case 12:
			goto tr516
		case 13:
			goto tr623
		case 32:
			goto tr961
		case 34:
			goto tr31
		case 44:
			goto tr962
		case 46:
			goto st655
		case 69:
			goto st207
		case 92:
			goto st76
		case 101:
			goto st207
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st658
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr961
		}
		goto st6
	st659:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof659
		}
	st_case_659:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr778
		case 12:
			goto tr909
		case 13:
			goto tr780
		case 32:
			goto tr966
		case 34:
			goto tr31
		case 44:
			goto tr967
		case 92:
			goto st76
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr966
		}
		goto st6
	st660:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof660
		}
	st_case_660:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr620
		case 12:
			goto tr516
		case 13:
			goto tr623
		case 32:
			goto tr961
		case 34:
			goto tr31
		case 44:
			goto tr962
		case 46:
			goto st655
		case 69:
			goto st207
		case 92:
			goto st76
		case 101:
			goto st207
		case 105:
			goto st659
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st660
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr961
		}
		goto st6
tr110:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st661
	st661:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof661
		}
	st_case_661:
//line plugins/parsers/influx/machine.go:26537
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr620
		case 12:
			goto tr516
		case 13:
			goto tr623
		case 32:
			goto tr961
		case 34:
			goto tr31
		case 44:
			goto tr962
		case 46:
			goto st655
		case 69:
			goto st207
		case 92:
			goto st76
		case 101:
			goto st207
		case 105:
			goto st659
		case 117:
			goto st662
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st658
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr961
		}
		goto st6
	st662:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof662
		}
	st_case_662:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr784
		case 12:
			goto tr912
		case 13:
			goto tr786
		case 32:
			goto tr969
		case 34:
			goto tr31
		case 44:
			goto tr970
		case 92:
			goto st76
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr969
		}
		goto st6
tr111:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st663
	st663:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof663
		}
	st_case_663:
//line plugins/parsers/influx/machine.go:26609
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr620
		case 12:
			goto tr516
		case 13:
			goto tr623
		case 32:
			goto tr961
		case 34:
			goto tr31
		case 44:
			goto tr962
		case 46:
			goto st655
		case 69:
			goto st207
		case 92:
			goto st76
		case 101:
			goto st207
		case 105:
			goto st659
		case 117:
			goto st662
		}
		switch {
		case ( m.data)[( m.p)] > 11:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st663
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr961
		}
		goto st6
tr112:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st664
	st664:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof664
		}
	st_case_664:
//line plugins/parsers/influx/machine.go:26656
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr790
		case 12:
			goto tr916
		case 13:
			goto tr792
		case 32:
			goto tr972
		case 34:
			goto tr31
		case 44:
			goto tr973
		case 65:
			goto st209
		case 92:
			goto st76
		case 97:
			goto st212
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr972
		}
		goto st6
	st209:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof209
		}
	st_case_209:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 76:
			goto st210
		case 92:
			goto st76
		}
		goto st6
	st210:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof210
		}
	st_case_210:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 83:
			goto st211
		case 92:
			goto st76
		}
		goto st6
	st211:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof211
		}
	st_case_211:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 69:
			goto st665
		case 92:
			goto st76
		}
		goto st6
	st665:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof665
		}
	st_case_665:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr790
		case 12:
			goto tr916
		case 13:
			goto tr792
		case 32:
			goto tr972
		case 34:
			goto tr31
		case 44:
			goto tr973
		case 92:
			goto st76
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr972
		}
		goto st6
	st212:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof212
		}
	st_case_212:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 92:
			goto st76
		case 108:
			goto st213
		}
		goto st6
	st213:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof213
		}
	st_case_213:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 92:
			goto st76
		case 115:
			goto st214
		}
		goto st6
	st214:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof214
		}
	st_case_214:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 92:
			goto st76
		case 101:
			goto st665
		}
		goto st6
tr113:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st666
	st666:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof666
		}
	st_case_666:
//line plugins/parsers/influx/machine.go:26837
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr790
		case 12:
			goto tr916
		case 13:
			goto tr792
		case 32:
			goto tr972
		case 34:
			goto tr31
		case 44:
			goto tr973
		case 82:
			goto st215
		case 92:
			goto st76
		case 114:
			goto st216
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr972
		}
		goto st6
	st215:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof215
		}
	st_case_215:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 85:
			goto st211
		case 92:
			goto st76
		}
		goto st6
	st216:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof216
		}
	st_case_216:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st7
		case 12:
			goto tr8
		case 13:
			goto st8
		case 34:
			goto tr31
		case 92:
			goto st76
		case 117:
			goto st214
		}
		goto st6
tr114:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st667
	st667:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof667
		}
	st_case_667:
//line plugins/parsers/influx/machine.go:26913
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr790
		case 12:
			goto tr916
		case 13:
			goto tr792
		case 32:
			goto tr972
		case 34:
			goto tr31
		case 44:
			goto tr973
		case 92:
			goto st76
		case 97:
			goto st212
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr972
		}
		goto st6
tr115:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st668
	st668:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof668
		}
	st_case_668:
//line plugins/parsers/influx/machine.go:26947
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr790
		case 12:
			goto tr916
		case 13:
			goto tr792
		case 32:
			goto tr972
		case 34:
			goto tr31
		case 44:
			goto tr973
		case 92:
			goto st76
		case 114:
			goto st216
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 11 {
			goto tr972
		}
		goto st6
tr378:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st217
	st217:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof217
		}
	st_case_217:
//line plugins/parsers/influx/machine.go:26981
		switch ( m.data)[( m.p)] {
		case 34:
			goto st203
		case 92:
			goto st203
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr8
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr8
		}
		goto st3
tr96:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st218
	st218:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof218
		}
	st_case_218:
//line plugins/parsers/influx/machine.go:27008
		switch ( m.data)[( m.p)] {
		case 9:
			goto st32
		case 10:
			goto st7
		case 11:
			goto tr96
		case 12:
			goto st2
		case 13:
			goto st8
		case 32:
			goto st32
		case 34:
			goto tr97
		case 44:
			goto st6
		case 61:
			goto tr101
		case 92:
			goto tr98
		}
		goto tr94
tr74:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st219
	st219:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof219
		}
	st_case_219:
//line plugins/parsers/influx/machine.go:27043
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr3
		case 13:
			goto tr47
		case 32:
			goto tr1
		case 44:
			goto tr4
		case 46:
			goto st220
		case 48:
			goto st670
		case 92:
			goto st96
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st673
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr1
		}
		goto st1
tr75:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st220
	st220:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof220
		}
	st_case_220:
//line plugins/parsers/influx/machine.go:27082
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr3
		case 13:
			goto tr47
		case 32:
			goto tr1
		case 44:
			goto tr4
		case 92:
			goto st96
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st669
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr1
		}
		goto st1
	st669:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof669
		}
	st_case_669:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 11:
			goto tr798
		case 13:
			goto tr717
		case 32:
			goto tr622
		case 44:
			goto tr799
		case 69:
			goto st221
		case 92:
			goto st96
		case 101:
			goto st221
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st669
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr622
		}
		goto st1
	st221:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof221
		}
	st_case_221:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr3
		case 13:
			goto tr47
		case 32:
			goto tr1
		case 34:
			goto st222
		case 44:
			goto tr4
		case 92:
			goto st96
		}
		switch {
		case ( m.data)[( m.p)] < 43:
			if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
				goto tr1
			}
		case ( m.data)[( m.p)] > 45:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st520
			}
		default:
			goto st222
		}
		goto st1
	st222:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof222
		}
	st_case_222:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr3
		case 13:
			goto tr47
		case 32:
			goto tr1
		case 44:
			goto tr4
		case 92:
			goto st96
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st520
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr1
		}
		goto st1
	st670:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof670
		}
	st_case_670:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 11:
			goto tr798
		case 13:
			goto tr717
		case 32:
			goto tr622
		case 44:
			goto tr799
		case 46:
			goto st669
		case 69:
			goto st221
		case 92:
			goto st96
		case 101:
			goto st221
		case 105:
			goto st672
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st671
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr622
		}
		goto st1
	st671:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof671
		}
	st_case_671:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 11:
			goto tr798
		case 13:
			goto tr717
		case 32:
			goto tr622
		case 44:
			goto tr799
		case 46:
			goto st669
		case 69:
			goto st221
		case 92:
			goto st96
		case 101:
			goto st221
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st671
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr622
		}
		goto st1
	st672:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof672
		}
	st_case_672:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr925
		case 11:
			goto tr981
		case 13:
			goto tr927
		case 32:
			goto tr804
		case 44:
			goto tr982
		case 92:
			goto st96
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr804
		}
		goto st1
	st673:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof673
		}
	st_case_673:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 11:
			goto tr798
		case 13:
			goto tr717
		case 32:
			goto tr622
		case 44:
			goto tr799
		case 46:
			goto st669
		case 69:
			goto st221
		case 92:
			goto st96
		case 101:
			goto st221
		case 105:
			goto st672
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st673
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr622
		}
		goto st1
tr76:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st674
	st674:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof674
		}
	st_case_674:
//line plugins/parsers/influx/machine.go:27340
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 11:
			goto tr798
		case 13:
			goto tr717
		case 32:
			goto tr622
		case 44:
			goto tr799
		case 46:
			goto st669
		case 69:
			goto st221
		case 92:
			goto st96
		case 101:
			goto st221
		case 105:
			goto st672
		case 117:
			goto st675
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st671
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr622
		}
		goto st1
	st675:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof675
		}
	st_case_675:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr930
		case 11:
			goto tr984
		case 13:
			goto tr932
		case 32:
			goto tr809
		case 44:
			goto tr985
		case 92:
			goto st96
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr809
		}
		goto st1
tr77:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st676
	st676:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof676
		}
	st_case_676:
//line plugins/parsers/influx/machine.go:27408
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 11:
			goto tr798
		case 13:
			goto tr717
		case 32:
			goto tr622
		case 44:
			goto tr799
		case 46:
			goto st669
		case 69:
			goto st221
		case 92:
			goto st96
		case 101:
			goto st221
		case 105:
			goto st672
		case 117:
			goto st675
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st676
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr622
		}
		goto st1
tr78:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st677
	st677:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof677
		}
	st_case_677:
//line plugins/parsers/influx/machine.go:27453
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr935
		case 11:
			goto tr987
		case 13:
			goto tr937
		case 32:
			goto tr814
		case 44:
			goto tr988
		case 65:
			goto st223
		case 92:
			goto st96
		case 97:
			goto st226
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr814
		}
		goto st1
	st223:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof223
		}
	st_case_223:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr3
		case 13:
			goto tr47
		case 32:
			goto tr1
		case 44:
			goto tr4
		case 76:
			goto st224
		case 92:
			goto st96
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr1
		}
		goto st1
	st224:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof224
		}
	st_case_224:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr3
		case 13:
			goto tr47
		case 32:
			goto tr1
		case 44:
			goto tr4
		case 83:
			goto st225
		case 92:
			goto st96
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr1
		}
		goto st1
	st225:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof225
		}
	st_case_225:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr3
		case 13:
			goto tr47
		case 32:
			goto tr1
		case 44:
			goto tr4
		case 69:
			goto st678
		case 92:
			goto st96
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr1
		}
		goto st1
	st678:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof678
		}
	st_case_678:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr935
		case 11:
			goto tr987
		case 13:
			goto tr937
		case 32:
			goto tr814
		case 44:
			goto tr988
		case 92:
			goto st96
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr814
		}
		goto st1
	st226:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof226
		}
	st_case_226:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr3
		case 13:
			goto tr47
		case 32:
			goto tr1
		case 44:
			goto tr4
		case 92:
			goto st96
		case 108:
			goto st227
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr1
		}
		goto st1
	st227:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof227
		}
	st_case_227:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr3
		case 13:
			goto tr47
		case 32:
			goto tr1
		case 44:
			goto tr4
		case 92:
			goto st96
		case 115:
			goto st228
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr1
		}
		goto st1
	st228:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof228
		}
	st_case_228:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr3
		case 13:
			goto tr47
		case 32:
			goto tr1
		case 44:
			goto tr4
		case 92:
			goto st96
		case 101:
			goto st678
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr1
		}
		goto st1
tr79:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st679
	st679:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof679
		}
	st_case_679:
//line plugins/parsers/influx/machine.go:27660
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr935
		case 11:
			goto tr987
		case 13:
			goto tr937
		case 32:
			goto tr814
		case 44:
			goto tr988
		case 82:
			goto st229
		case 92:
			goto st96
		case 114:
			goto st230
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr814
		}
		goto st1
	st229:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof229
		}
	st_case_229:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr3
		case 13:
			goto tr47
		case 32:
			goto tr1
		case 44:
			goto tr4
		case 85:
			goto st225
		case 92:
			goto st96
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr1
		}
		goto st1
	st230:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof230
		}
	st_case_230:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr47
		case 11:
			goto tr3
		case 13:
			goto tr47
		case 32:
			goto tr1
		case 44:
			goto tr4
		case 92:
			goto st96
		case 117:
			goto st228
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr1
		}
		goto st1
tr80:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st680
	st680:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof680
		}
	st_case_680:
//line plugins/parsers/influx/machine.go:27744
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr935
		case 11:
			goto tr987
		case 13:
			goto tr937
		case 32:
			goto tr814
		case 44:
			goto tr988
		case 92:
			goto st96
		case 97:
			goto st226
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr814
		}
		goto st1
tr81:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st681
	st681:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof681
		}
	st_case_681:
//line plugins/parsers/influx/machine.go:27776
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr935
		case 11:
			goto tr987
		case 13:
			goto tr937
		case 32:
			goto tr814
		case 44:
			goto tr988
		case 92:
			goto st96
		case 114:
			goto st230
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr814
		}
		goto st1
tr44:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st231
tr405:
	( m.cs) = 231
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st231:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof231
		}
	st_case_231:
//line plugins/parsers/influx/machine.go:27825
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr404
		case 11:
			goto tr405
		case 13:
			goto tr404
		case 32:
			goto tr38
		case 44:
			goto tr4
		case 61:
			goto tr406
		case 92:
			goto tr45
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr38
		}
		goto tr41
tr40:
	( m.cs) = 232
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st232:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof232
		}
	st_case_232:
//line plugins/parsers/influx/machine.go:27868
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr404
		case 11:
			goto tr405
		case 13:
			goto tr404
		case 32:
			goto tr38
		case 44:
			goto tr4
		case 61:
			goto tr33
		case 92:
			goto tr45
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr38
		}
		goto tr41
tr445:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st233
	st233:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof233
		}
	st_case_233:
//line plugins/parsers/influx/machine.go:27900
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st682
		}
		goto tr407
tr446:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st682
	st682:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof682
		}
	st_case_682:
//line plugins/parsers/influx/machine.go:27916
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 13:
			goto tr453
		case 32:
			goto tr450
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st683
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto tr407
	st683:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof683
		}
	st_case_683:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 13:
			goto tr453
		case 32:
			goto tr450
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st684
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto tr407
	st684:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof684
		}
	st_case_684:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 13:
			goto tr453
		case 32:
			goto tr450
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st685
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto tr407
	st685:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof685
		}
	st_case_685:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 13:
			goto tr453
		case 32:
			goto tr450
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st686
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto tr407
	st686:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof686
		}
	st_case_686:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 13:
			goto tr453
		case 32:
			goto tr450
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st687
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto tr407
	st687:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof687
		}
	st_case_687:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 13:
			goto tr453
		case 32:
			goto tr450
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st688
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto tr407
	st688:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof688
		}
	st_case_688:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 13:
			goto tr453
		case 32:
			goto tr450
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st689
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto tr407
	st689:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof689
		}
	st_case_689:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 13:
			goto tr453
		case 32:
			goto tr450
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st690
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto tr407
	st690:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof690
		}
	st_case_690:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 13:
			goto tr453
		case 32:
			goto tr450
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st691
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto tr407
	st691:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof691
		}
	st_case_691:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 13:
			goto tr453
		case 32:
			goto tr450
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st692
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto tr407
	st692:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof692
		}
	st_case_692:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 13:
			goto tr453
		case 32:
			goto tr450
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st693
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto tr407
	st693:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof693
		}
	st_case_693:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 13:
			goto tr453
		case 32:
			goto tr450
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st694
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto tr407
	st694:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof694
		}
	st_case_694:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 13:
			goto tr453
		case 32:
			goto tr450
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st695
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto tr407
	st695:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof695
		}
	st_case_695:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 13:
			goto tr453
		case 32:
			goto tr450
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st696
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto tr407
	st696:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof696
		}
	st_case_696:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 13:
			goto tr453
		case 32:
			goto tr450
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st697
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto tr407
	st697:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof697
		}
	st_case_697:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 13:
			goto tr453
		case 32:
			goto tr450
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st698
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto tr407
	st698:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof698
		}
	st_case_698:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 13:
			goto tr453
		case 32:
			goto tr450
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st699
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto tr407
	st699:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof699
		}
	st_case_699:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 13:
			goto tr453
		case 32:
			goto tr450
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st700
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr450
		}
		goto tr407
	st700:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof700
		}
	st_case_700:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr451
		case 13:
			goto tr453
		case 32:
			goto tr450
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr450
		}
		goto tr407
tr15:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st234
	st234:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof234
		}
	st_case_234:
//line plugins/parsers/influx/machine.go:28336
		switch ( m.data)[( m.p)] {
		case 46:
			goto st235
		case 48:
			goto st702
		}
		if 49 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st705
		}
		goto tr8
tr16:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st235
	st235:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof235
		}
	st_case_235:
//line plugins/parsers/influx/machine.go:28358
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st701
		}
		goto tr8
	st701:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof701
		}
	st_case_701:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 13:
			goto tr717
		case 32:
			goto tr516
		case 44:
			goto tr907
		case 69:
			goto st236
		case 101:
			goto st236
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st701
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr516
		}
		goto tr105
	st236:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof236
		}
	st_case_236:
		switch ( m.data)[( m.p)] {
		case 34:
			goto st237
		case 43:
			goto st237
		case 45:
			goto st237
		}
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st611
		}
		goto tr8
	st237:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof237
		}
	st_case_237:
		if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
			goto st611
		}
		goto tr8
	st702:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof702
		}
	st_case_702:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 13:
			goto tr717
		case 32:
			goto tr516
		case 44:
			goto tr907
		case 46:
			goto st701
		case 69:
			goto st236
		case 101:
			goto st236
		case 105:
			goto st704
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st703
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr516
		}
		goto tr105
	st703:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof703
		}
	st_case_703:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 13:
			goto tr717
		case 32:
			goto tr516
		case 44:
			goto tr907
		case 46:
			goto st701
		case 69:
			goto st236
		case 101:
			goto st236
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st703
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr516
		}
		goto tr105
	st704:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof704
		}
	st_case_704:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr925
		case 13:
			goto tr927
		case 32:
			goto tr909
		case 44:
			goto tr1014
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr909
		}
		goto tr105
	st705:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof705
		}
	st_case_705:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 13:
			goto tr717
		case 32:
			goto tr516
		case 44:
			goto tr907
		case 46:
			goto st701
		case 69:
			goto st236
		case 101:
			goto st236
		case 105:
			goto st704
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st705
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr516
		}
		goto tr105
tr17:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st706
	st706:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof706
		}
	st_case_706:
//line plugins/parsers/influx/machine.go:28541
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 13:
			goto tr717
		case 32:
			goto tr516
		case 44:
			goto tr907
		case 46:
			goto st701
		case 69:
			goto st236
		case 101:
			goto st236
		case 105:
			goto st704
		case 117:
			goto st707
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st703
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr516
		}
		goto tr105
	st707:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof707
		}
	st_case_707:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr930
		case 13:
			goto tr932
		case 32:
			goto tr912
		case 44:
			goto tr1016
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr912
		}
		goto tr105
tr18:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st708
	st708:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof708
		}
	st_case_708:
//line plugins/parsers/influx/machine.go:28601
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr715
		case 13:
			goto tr717
		case 32:
			goto tr516
		case 44:
			goto tr907
		case 46:
			goto st701
		case 69:
			goto st236
		case 101:
			goto st236
		case 105:
			goto st704
		case 117:
			goto st707
		}
		switch {
		case ( m.data)[( m.p)] > 12:
			if 48 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 57 {
				goto st708
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr516
		}
		goto tr105
tr19:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st709
	st709:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof709
		}
	st_case_709:
//line plugins/parsers/influx/machine.go:28642
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr935
		case 13:
			goto tr937
		case 32:
			goto tr916
		case 44:
			goto tr1018
		case 65:
			goto st238
		case 97:
			goto st241
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr916
		}
		goto tr105
	st238:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof238
		}
	st_case_238:
		if ( m.data)[( m.p)] == 76 {
			goto st239
		}
		goto tr8
	st239:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof239
		}
	st_case_239:
		if ( m.data)[( m.p)] == 83 {
			goto st240
		}
		goto tr8
	st240:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof240
		}
	st_case_240:
		if ( m.data)[( m.p)] == 69 {
			goto st710
		}
		goto tr8
	st710:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof710
		}
	st_case_710:
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr935
		case 13:
			goto tr937
		case 32:
			goto tr916
		case 44:
			goto tr1018
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr916
		}
		goto tr105
	st241:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof241
		}
	st_case_241:
		if ( m.data)[( m.p)] == 108 {
			goto st242
		}
		goto tr8
	st242:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof242
		}
	st_case_242:
		if ( m.data)[( m.p)] == 115 {
			goto st243
		}
		goto tr8
	st243:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof243
		}
	st_case_243:
		if ( m.data)[( m.p)] == 101 {
			goto st710
		}
		goto tr8
tr20:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st711
	st711:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof711
		}
	st_case_711:
//line plugins/parsers/influx/machine.go:28745
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr935
		case 13:
			goto tr937
		case 32:
			goto tr916
		case 44:
			goto tr1018
		case 82:
			goto st244
		case 114:
			goto st245
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr916
		}
		goto tr105
	st244:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof244
		}
	st_case_244:
		if ( m.data)[( m.p)] == 85 {
			goto st240
		}
		goto tr8
	st245:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof245
		}
	st_case_245:
		if ( m.data)[( m.p)] == 117 {
			goto st243
		}
		goto tr8
tr21:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st712
	st712:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof712
		}
	st_case_712:
//line plugins/parsers/influx/machine.go:28793
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr935
		case 13:
			goto tr937
		case 32:
			goto tr916
		case 44:
			goto tr1018
		case 97:
			goto st241
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr916
		}
		goto tr105
tr22:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st713
	st713:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof713
		}
	st_case_713:
//line plugins/parsers/influx/machine.go:28821
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr935
		case 13:
			goto tr937
		case 32:
			goto tr916
		case 44:
			goto tr1018
		case 114:
			goto st245
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto tr916
		}
		goto tr105
tr9:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st246
	st246:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof246
		}
	st_case_246:
//line plugins/parsers/influx/machine.go:28849
		switch ( m.data)[( m.p)] {
		case 10:
			goto tr8
		case 11:
			goto tr9
		case 13:
			goto tr8
		case 32:
			goto st2
		case 44:
			goto tr8
		case 61:
			goto tr12
		case 92:
			goto tr10
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st2
		}
		goto tr6
	st247:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof247
		}
	st_case_247:
		if ( m.data)[( m.p)] == 10 {
			goto tr421
		}
		goto st247
tr421:
//line plugins/parsers/influx/machine.go.rl:69

	{goto st715 }

	goto st714
	st714:
//line plugins/parsers/influx/machine.go.rl:157

	m.lineno++
	m.sol = m.p
	m.sol++ // next char will be the first column in the line

		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof714
		}
	st_case_714:
//line plugins/parsers/influx/machine.go:28896
		goto st0
	st250:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof250
		}
	st_case_250:
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr35
		case 35:
			goto tr35
		case 44:
			goto tr35
		case 92:
			goto tr425
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr35
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr35
		}
		goto tr424
tr424:
//line plugins/parsers/influx/machine.go.rl:73

	foundMetric = true

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st717
	st717:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof717
		}
	st_case_717:
//line plugins/parsers/influx/machine.go:28937
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr2
		case 10:
			goto tr1026
		case 12:
			goto tr2
		case 13:
			goto tr1027
		case 32:
			goto tr2
		case 44:
			goto tr1028
		case 92:
			goto st258
		}
		goto st717
tr1026:
	( m.cs) = 718
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr1030:
	( m.cs) = 718
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st718:
//line plugins/parsers/influx/machine.go.rl:157

	m.lineno++
	m.sol = m.p
	m.sol++ // next char will be the first column in the line

//line plugins/parsers/influx/machine.go.rl:163

	( m.cs) = 715;
	{( m.p)++; goto _out }

		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof718
		}
	st_case_718:
//line plugins/parsers/influx/machine.go:28997
		goto st0
tr1027:
	( m.cs) = 251
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr1031:
	( m.cs) = 251
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st251:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof251
		}
	st_case_251:
//line plugins/parsers/influx/machine.go:29030
		if ( m.data)[( m.p)] == 10 {
			goto st718
		}
		goto st0
tr1028:
	( m.cs) = 252
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
tr1032:
	( m.cs) = 252
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; goto _out }
	}

	goto _again
	st252:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof252
		}
	st_case_252:
//line plugins/parsers/influx/machine.go:29066
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr2
		case 44:
			goto tr2
		case 61:
			goto tr2
		case 92:
			goto tr428
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr2
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr2
		}
		goto tr427
tr427:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st253
	st253:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof253
		}
	st_case_253:
//line plugins/parsers/influx/machine.go:29097
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr2
		case 44:
			goto tr2
		case 61:
			goto tr430
		case 92:
			goto st256
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr2
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr2
		}
		goto st253
tr430:
//line plugins/parsers/influx/machine.go.rl:86

	key = m.text()

	goto st254
	st254:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof254
		}
	st_case_254:
//line plugins/parsers/influx/machine.go:29128
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr2
		case 44:
			goto tr2
		case 61:
			goto tr2
		case 92:
			goto tr433
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr2
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr2
		}
		goto tr432
tr432:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st719
	st719:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof719
		}
	st_case_719:
//line plugins/parsers/influx/machine.go:29159
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr2
		case 10:
			goto tr1030
		case 12:
			goto tr2
		case 13:
			goto tr1031
		case 32:
			goto tr2
		case 44:
			goto tr1032
		case 61:
			goto tr2
		case 92:
			goto st255
		}
		goto st719
tr433:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st255
	st255:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof255
		}
	st_case_255:
//line plugins/parsers/influx/machine.go:29190
		if ( m.data)[( m.p)] == 92 {
			goto st720
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr2
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr2
		}
		goto st719
	st720:
//line plugins/parsers/influx/machine.go.rl:234
 ( m.p)--
 
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof720
		}
	st_case_720:
//line plugins/parsers/influx/machine.go:29211
		switch ( m.data)[( m.p)] {
		case 9:
			goto tr2
		case 10:
			goto tr1030
		case 12:
			goto tr2
		case 13:
			goto tr1031
		case 32:
			goto tr2
		case 44:
			goto tr1032
		case 61:
			goto tr2
		case 92:
			goto st255
		}
		goto st719
tr428:
//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st256
	st256:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof256
		}
	st_case_256:
//line plugins/parsers/influx/machine.go:29242
		if ( m.data)[( m.p)] == 92 {
			goto st257
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr2
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr2
		}
		goto st253
	st257:
//line plugins/parsers/influx/machine.go.rl:234
 ( m.p)--
 
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof257
		}
	st_case_257:
//line plugins/parsers/influx/machine.go:29263
		switch ( m.data)[( m.p)] {
		case 32:
			goto tr2
		case 44:
			goto tr2
		case 61:
			goto tr430
		case 92:
			goto st256
		}
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto tr2
			}
		case ( m.data)[( m.p)] >= 9:
			goto tr2
		}
		goto st253
tr425:
//line plugins/parsers/influx/machine.go.rl:73

	foundMetric = true

//line plugins/parsers/influx/machine.go.rl:19

	m.pb = m.p

	goto st258
	st258:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof258
		}
	st_case_258:
//line plugins/parsers/influx/machine.go:29298
		switch {
		case ( m.data)[( m.p)] > 10:
			if 12 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 13 {
				goto st0
			}
		case ( m.data)[( m.p)] >= 9:
			goto st0
		}
		goto st717
	st715:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof715
		}
	st_case_715:
		switch ( m.data)[( m.p)] {
		case 10:
			goto st716
		case 13:
			goto st248
		case 32:
			goto st715
		case 35:
			goto st249
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st715
		}
		goto tr1023
	st716:
//line plugins/parsers/influx/machine.go.rl:157

	m.lineno++
	m.sol = m.p
	m.sol++ // next char will be the first column in the line

		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof716
		}
	st_case_716:
//line plugins/parsers/influx/machine.go:29338
		switch ( m.data)[( m.p)] {
		case 10:
			goto st716
		case 13:
			goto st248
		case 32:
			goto st715
		case 35:
			goto st249
		}
		if 9 <= ( m.data)[( m.p)] && ( m.data)[( m.p)] <= 12 {
			goto st715
		}
		goto tr1023
	st248:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof248
		}
	st_case_248:
		if ( m.data)[( m.p)] == 10 {
			goto st716
		}
		goto st0
	st249:
		if ( m.p)++; ( m.p) == ( m.pe) {
			goto _test_eof249
		}
	st_case_249:
		if ( m.data)[( m.p)] == 10 {
			goto st716
		}
		goto st249
	st_out:
	_test_eof259: ( m.cs) = 259; goto _test_eof
	_test_eof1: ( m.cs) = 1; goto _test_eof
	_test_eof2: ( m.cs) = 2; goto _test_eof
	_test_eof3: ( m.cs) = 3; goto _test_eof
	_test_eof4: ( m.cs) = 4; goto _test_eof
	_test_eof5: ( m.cs) = 5; goto _test_eof
	_test_eof6: ( m.cs) = 6; goto _test_eof
	_test_eof7: ( m.cs) = 7; goto _test_eof
	_test_eof8: ( m.cs) = 8; goto _test_eof
	_test_eof260: ( m.cs) = 260; goto _test_eof
	_test_eof261: ( m.cs) = 261; goto _test_eof
	_test_eof262: ( m.cs) = 262; goto _test_eof
	_test_eof9: ( m.cs) = 9; goto _test_eof
	_test_eof10: ( m.cs) = 10; goto _test_eof
	_test_eof11: ( m.cs) = 11; goto _test_eof
	_test_eof12: ( m.cs) = 12; goto _test_eof
	_test_eof13: ( m.cs) = 13; goto _test_eof
	_test_eof14: ( m.cs) = 14; goto _test_eof
	_test_eof15: ( m.cs) = 15; goto _test_eof
	_test_eof16: ( m.cs) = 16; goto _test_eof
	_test_eof17: ( m.cs) = 17; goto _test_eof
	_test_eof18: ( m.cs) = 18; goto _test_eof
	_test_eof19: ( m.cs) = 19; goto _test_eof
	_test_eof20: ( m.cs) = 20; goto _test_eof
	_test_eof21: ( m.cs) = 21; goto _test_eof
	_test_eof22: ( m.cs) = 22; goto _test_eof
	_test_eof23: ( m.cs) = 23; goto _test_eof
	_test_eof24: ( m.cs) = 24; goto _test_eof
	_test_eof25: ( m.cs) = 25; goto _test_eof
	_test_eof26: ( m.cs) = 26; goto _test_eof
	_test_eof27: ( m.cs) = 27; goto _test_eof
	_test_eof28: ( m.cs) = 28; goto _test_eof
	_test_eof29: ( m.cs) = 29; goto _test_eof
	_test_eof30: ( m.cs) = 30; goto _test_eof
	_test_eof31: ( m.cs) = 31; goto _test_eof
	_test_eof32: ( m.cs) = 32; goto _test_eof
	_test_eof33: ( m.cs) = 33; goto _test_eof
	_test_eof263: ( m.cs) = 263; goto _test_eof
	_test_eof264: ( m.cs) = 264; goto _test_eof
	_test_eof34: ( m.cs) = 34; goto _test_eof
	_test_eof35: ( m.cs) = 35; goto _test_eof
	_test_eof265: ( m.cs) = 265; goto _test_eof
	_test_eof266: ( m.cs) = 266; goto _test_eof
	_test_eof267: ( m.cs) = 267; goto _test_eof
	_test_eof36: ( m.cs) = 36; goto _test_eof
	_test_eof268: ( m.cs) = 268; goto _test_eof
	_test_eof269: ( m.cs) = 269; goto _test_eof
	_test_eof270: ( m.cs) = 270; goto _test_eof
	_test_eof271: ( m.cs) = 271; goto _test_eof
	_test_eof272: ( m.cs) = 272; goto _test_eof
	_test_eof273: ( m.cs) = 273; goto _test_eof
	_test_eof274: ( m.cs) = 274; goto _test_eof
	_test_eof275: ( m.cs) = 275; goto _test_eof
	_test_eof276: ( m.cs) = 276; goto _test_eof
	_test_eof277: ( m.cs) = 277; goto _test_eof
	_test_eof278: ( m.cs) = 278; goto _test_eof
	_test_eof279: ( m.cs) = 279; goto _test_eof
	_test_eof280: ( m.cs) = 280; goto _test_eof
	_test_eof281: ( m.cs) = 281; goto _test_eof
	_test_eof282: ( m.cs) = 282; goto _test_eof
	_test_eof283: ( m.cs) = 283; goto _test_eof
	_test_eof284: ( m.cs) = 284; goto _test_eof
	_test_eof285: ( m.cs) = 285; goto _test_eof
	_test_eof37: ( m.cs) = 37; goto _test_eof
	_test_eof38: ( m.cs) = 38; goto _test_eof
	_test_eof286: ( m.cs) = 286; goto _test_eof
	_test_eof287: ( m.cs) = 287; goto _test_eof
	_test_eof288: ( m.cs) = 288; goto _test_eof
	_test_eof39: ( m.cs) = 39; goto _test_eof
	_test_eof40: ( m.cs) = 40; goto _test_eof
	_test_eof41: ( m.cs) = 41; goto _test_eof
	_test_eof42: ( m.cs) = 42; goto _test_eof
	_test_eof43: ( m.cs) = 43; goto _test_eof
	_test_eof289: ( m.cs) = 289; goto _test_eof
	_test_eof290: ( m.cs) = 290; goto _test_eof
	_test_eof291: ( m.cs) = 291; goto _test_eof
	_test_eof292: ( m.cs) = 292; goto _test_eof
	_test_eof44: ( m.cs) = 44; goto _test_eof
	_test_eof293: ( m.cs) = 293; goto _test_eof
	_test_eof294: ( m.cs) = 294; goto _test_eof
	_test_eof295: ( m.cs) = 295; goto _test_eof
	_test_eof296: ( m.cs) = 296; goto _test_eof
	_test_eof297: ( m.cs) = 297; goto _test_eof
	_test_eof298: ( m.cs) = 298; goto _test_eof
	_test_eof299: ( m.cs) = 299; goto _test_eof
	_test_eof300: ( m.cs) = 300; goto _test_eof
	_test_eof301: ( m.cs) = 301; goto _test_eof
	_test_eof302: ( m.cs) = 302; goto _test_eof
	_test_eof303: ( m.cs) = 303; goto _test_eof
	_test_eof304: ( m.cs) = 304; goto _test_eof
	_test_eof305: ( m.cs) = 305; goto _test_eof
	_test_eof306: ( m.cs) = 306; goto _test_eof
	_test_eof307: ( m.cs) = 307; goto _test_eof
	_test_eof308: ( m.cs) = 308; goto _test_eof
	_test_eof309: ( m.cs) = 309; goto _test_eof
	_test_eof310: ( m.cs) = 310; goto _test_eof
	_test_eof311: ( m.cs) = 311; goto _test_eof
	_test_eof312: ( m.cs) = 312; goto _test_eof
	_test_eof313: ( m.cs) = 313; goto _test_eof
	_test_eof314: ( m.cs) = 314; goto _test_eof
	_test_eof45: ( m.cs) = 45; goto _test_eof
	_test_eof46: ( m.cs) = 46; goto _test_eof
	_test_eof47: ( m.cs) = 47; goto _test_eof
	_test_eof48: ( m.cs) = 48; goto _test_eof
	_test_eof49: ( m.cs) = 49; goto _test_eof
	_test_eof50: ( m.cs) = 50; goto _test_eof
	_test_eof51: ( m.cs) = 51; goto _test_eof
	_test_eof52: ( m.cs) = 52; goto _test_eof
	_test_eof53: ( m.cs) = 53; goto _test_eof
	_test_eof54: ( m.cs) = 54; goto _test_eof
	_test_eof315: ( m.cs) = 315; goto _test_eof
	_test_eof316: ( m.cs) = 316; goto _test_eof
	_test_eof317: ( m.cs) = 317; goto _test_eof
	_test_eof55: ( m.cs) = 55; goto _test_eof
	_test_eof56: ( m.cs) = 56; goto _test_eof
	_test_eof57: ( m.cs) = 57; goto _test_eof
	_test_eof58: ( m.cs) = 58; goto _test_eof
	_test_eof59: ( m.cs) = 59; goto _test_eof
	_test_eof60: ( m.cs) = 60; goto _test_eof
	_test_eof318: ( m.cs) = 318; goto _test_eof
	_test_eof319: ( m.cs) = 319; goto _test_eof
	_test_eof61: ( m.cs) = 61; goto _test_eof
	_test_eof320: ( m.cs) = 320; goto _test_eof
	_test_eof321: ( m.cs) = 321; goto _test_eof
	_test_eof322: ( m.cs) = 322; goto _test_eof
	_test_eof323: ( m.cs) = 323; goto _test_eof
	_test_eof324: ( m.cs) = 324; goto _test_eof
	_test_eof325: ( m.cs) = 325; goto _test_eof
	_test_eof326: ( m.cs) = 326; goto _test_eof
	_test_eof327: ( m.cs) = 327; goto _test_eof
	_test_eof328: ( m.cs) = 328; goto _test_eof
	_test_eof329: ( m.cs) = 329; goto _test_eof
	_test_eof330: ( m.cs) = 330; goto _test_eof
	_test_eof331: ( m.cs) = 331; goto _test_eof
	_test_eof332: ( m.cs) = 332; goto _test_eof
	_test_eof333: ( m.cs) = 333; goto _test_eof
	_test_eof334: ( m.cs) = 334; goto _test_eof
	_test_eof335: ( m.cs) = 335; goto _test_eof
	_test_eof336: ( m.cs) = 336; goto _test_eof
	_test_eof337: ( m.cs) = 337; goto _test_eof
	_test_eof338: ( m.cs) = 338; goto _test_eof
	_test_eof339: ( m.cs) = 339; goto _test_eof
	_test_eof62: ( m.cs) = 62; goto _test_eof
	_test_eof340: ( m.cs) = 340; goto _test_eof
	_test_eof341: ( m.cs) = 341; goto _test_eof
	_test_eof342: ( m.cs) = 342; goto _test_eof
	_test_eof63: ( m.cs) = 63; goto _test_eof
	_test_eof343: ( m.cs) = 343; goto _test_eof
	_test_eof344: ( m.cs) = 344; goto _test_eof
	_test_eof345: ( m.cs) = 345; goto _test_eof
	_test_eof346: ( m.cs) = 346; goto _test_eof
	_test_eof347: ( m.cs) = 347; goto _test_eof
	_test_eof348: ( m.cs) = 348; goto _test_eof
	_test_eof349: ( m.cs) = 349; goto _test_eof
	_test_eof350: ( m.cs) = 350; goto _test_eof
	_test_eof351: ( m.cs) = 351; goto _test_eof
	_test_eof352: ( m.cs) = 352; goto _test_eof
	_test_eof353: ( m.cs) = 353; goto _test_eof
	_test_eof354: ( m.cs) = 354; goto _test_eof
	_test_eof355: ( m.cs) = 355; goto _test_eof
	_test_eof356: ( m.cs) = 356; goto _test_eof
	_test_eof357: ( m.cs) = 357; goto _test_eof
	_test_eof358: ( m.cs) = 358; goto _test_eof
	_test_eof359: ( m.cs) = 359; goto _test_eof
	_test_eof360: ( m.cs) = 360; goto _test_eof
	_test_eof361: ( m.cs) = 361; goto _test_eof
	_test_eof362: ( m.cs) = 362; goto _test_eof
	_test_eof64: ( m.cs) = 64; goto _test_eof
	_test_eof65: ( m.cs) = 65; goto _test_eof
	_test_eof66: ( m.cs) = 66; goto _test_eof
	_test_eof67: ( m.cs) = 67; goto _test_eof
	_test_eof68: ( m.cs) = 68; goto _test_eof
	_test_eof363: ( m.cs) = 363; goto _test_eof
	_test_eof69: ( m.cs) = 69; goto _test_eof
	_test_eof70: ( m.cs) = 70; goto _test_eof
	_test_eof71: ( m.cs) = 71; goto _test_eof
	_test_eof72: ( m.cs) = 72; goto _test_eof
	_test_eof73: ( m.cs) = 73; goto _test_eof
	_test_eof364: ( m.cs) = 364; goto _test_eof
	_test_eof365: ( m.cs) = 365; goto _test_eof
	_test_eof366: ( m.cs) = 366; goto _test_eof
	_test_eof74: ( m.cs) = 74; goto _test_eof
	_test_eof75: ( m.cs) = 75; goto _test_eof
	_test_eof367: ( m.cs) = 367; goto _test_eof
	_test_eof368: ( m.cs) = 368; goto _test_eof
	_test_eof76: ( m.cs) = 76; goto _test_eof
	_test_eof369: ( m.cs) = 369; goto _test_eof
	_test_eof77: ( m.cs) = 77; goto _test_eof
	_test_eof370: ( m.cs) = 370; goto _test_eof
	_test_eof371: ( m.cs) = 371; goto _test_eof
	_test_eof372: ( m.cs) = 372; goto _test_eof
	_test_eof373: ( m.cs) = 373; goto _test_eof
	_test_eof374: ( m.cs) = 374; goto _test_eof
	_test_eof375: ( m.cs) = 375; goto _test_eof
	_test_eof376: ( m.cs) = 376; goto _test_eof
	_test_eof377: ( m.cs) = 377; goto _test_eof
	_test_eof378: ( m.cs) = 378; goto _test_eof
	_test_eof379: ( m.cs) = 379; goto _test_eof
	_test_eof380: ( m.cs) = 380; goto _test_eof
	_test_eof381: ( m.cs) = 381; goto _test_eof
	_test_eof382: ( m.cs) = 382; goto _test_eof
	_test_eof383: ( m.cs) = 383; goto _test_eof
	_test_eof384: ( m.cs) = 384; goto _test_eof
	_test_eof385: ( m.cs) = 385; goto _test_eof
	_test_eof386: ( m.cs) = 386; goto _test_eof
	_test_eof387: ( m.cs) = 387; goto _test_eof
	_test_eof388: ( m.cs) = 388; goto _test_eof
	_test_eof389: ( m.cs) = 389; goto _test_eof
	_test_eof78: ( m.cs) = 78; goto _test_eof
	_test_eof79: ( m.cs) = 79; goto _test_eof
	_test_eof80: ( m.cs) = 80; goto _test_eof
	_test_eof81: ( m.cs) = 81; goto _test_eof
	_test_eof82: ( m.cs) = 82; goto _test_eof
	_test_eof83: ( m.cs) = 83; goto _test_eof
	_test_eof84: ( m.cs) = 84; goto _test_eof
	_test_eof85: ( m.cs) = 85; goto _test_eof
	_test_eof86: ( m.cs) = 86; goto _test_eof
	_test_eof87: ( m.cs) = 87; goto _test_eof
	_test_eof88: ( m.cs) = 88; goto _test_eof
	_test_eof89: ( m.cs) = 89; goto _test_eof
	_test_eof90: ( m.cs) = 90; goto _test_eof
	_test_eof91: ( m.cs) = 91; goto _test_eof
	_test_eof390: ( m.cs) = 390; goto _test_eof
	_test_eof391: ( m.cs) = 391; goto _test_eof
	_test_eof392: ( m.cs) = 392; goto _test_eof
	_test_eof393: ( m.cs) = 393; goto _test_eof
	_test_eof92: ( m.cs) = 92; goto _test_eof
	_test_eof93: ( m.cs) = 93; goto _test_eof
	_test_eof94: ( m.cs) = 94; goto _test_eof
	_test_eof95: ( m.cs) = 95; goto _test_eof
	_test_eof394: ( m.cs) = 394; goto _test_eof
	_test_eof395: ( m.cs) = 395; goto _test_eof
	_test_eof96: ( m.cs) = 96; goto _test_eof
	_test_eof97: ( m.cs) = 97; goto _test_eof
	_test_eof396: ( m.cs) = 396; goto _test_eof
	_test_eof98: ( m.cs) = 98; goto _test_eof
	_test_eof99: ( m.cs) = 99; goto _test_eof
	_test_eof397: ( m.cs) = 397; goto _test_eof
	_test_eof398: ( m.cs) = 398; goto _test_eof
	_test_eof100: ( m.cs) = 100; goto _test_eof
	_test_eof399: ( m.cs) = 399; goto _test_eof
	_test_eof400: ( m.cs) = 400; goto _test_eof
	_test_eof101: ( m.cs) = 101; goto _test_eof
	_test_eof102: ( m.cs) = 102; goto _test_eof
	_test_eof401: ( m.cs) = 401; goto _test_eof
	_test_eof402: ( m.cs) = 402; goto _test_eof
	_test_eof403: ( m.cs) = 403; goto _test_eof
	_test_eof404: ( m.cs) = 404; goto _test_eof
	_test_eof405: ( m.cs) = 405; goto _test_eof
	_test_eof406: ( m.cs) = 406; goto _test_eof
	_test_eof407: ( m.cs) = 407; goto _test_eof
	_test_eof408: ( m.cs) = 408; goto _test_eof
	_test_eof409: ( m.cs) = 409; goto _test_eof
	_test_eof410: ( m.cs) = 410; goto _test_eof
	_test_eof411: ( m.cs) = 411; goto _test_eof
	_test_eof412: ( m.cs) = 412; goto _test_eof
	_test_eof413: ( m.cs) = 413; goto _test_eof
	_test_eof414: ( m.cs) = 414; goto _test_eof
	_test_eof415: ( m.cs) = 415; goto _test_eof
	_test_eof416: ( m.cs) = 416; goto _test_eof
	_test_eof417: ( m.cs) = 417; goto _test_eof
	_test_eof418: ( m.cs) = 418; goto _test_eof
	_test_eof103: ( m.cs) = 103; goto _test_eof
	_test_eof419: ( m.cs) = 419; goto _test_eof
	_test_eof420: ( m.cs) = 420; goto _test_eof
	_test_eof421: ( m.cs) = 421; goto _test_eof
	_test_eof104: ( m.cs) = 104; goto _test_eof
	_test_eof105: ( m.cs) = 105; goto _test_eof
	_test_eof422: ( m.cs) = 422; goto _test_eof
	_test_eof423: ( m.cs) = 423; goto _test_eof
	_test_eof424: ( m.cs) = 424; goto _test_eof
	_test_eof106: ( m.cs) = 106; goto _test_eof
	_test_eof425: ( m.cs) = 425; goto _test_eof
	_test_eof426: ( m.cs) = 426; goto _test_eof
	_test_eof427: ( m.cs) = 427; goto _test_eof
	_test_eof428: ( m.cs) = 428; goto _test_eof
	_test_eof429: ( m.cs) = 429; goto _test_eof
	_test_eof430: ( m.cs) = 430; goto _test_eof
	_test_eof431: ( m.cs) = 431; goto _test_eof
	_test_eof432: ( m.cs) = 432; goto _test_eof
	_test_eof433: ( m.cs) = 433; goto _test_eof
	_test_eof434: ( m.cs) = 434; goto _test_eof
	_test_eof435: ( m.cs) = 435; goto _test_eof
	_test_eof436: ( m.cs) = 436; goto _test_eof
	_test_eof437: ( m.cs) = 437; goto _test_eof
	_test_eof438: ( m.cs) = 438; goto _test_eof
	_test_eof439: ( m.cs) = 439; goto _test_eof
	_test_eof440: ( m.cs) = 440; goto _test_eof
	_test_eof441: ( m.cs) = 441; goto _test_eof
	_test_eof442: ( m.cs) = 442; goto _test_eof
	_test_eof443: ( m.cs) = 443; goto _test_eof
	_test_eof444: ( m.cs) = 444; goto _test_eof
	_test_eof107: ( m.cs) = 107; goto _test_eof
	_test_eof445: ( m.cs) = 445; goto _test_eof
	_test_eof446: ( m.cs) = 446; goto _test_eof
	_test_eof447: ( m.cs) = 447; goto _test_eof
	_test_eof448: ( m.cs) = 448; goto _test_eof
	_test_eof449: ( m.cs) = 449; goto _test_eof
	_test_eof450: ( m.cs) = 450; goto _test_eof
	_test_eof451: ( m.cs) = 451; goto _test_eof
	_test_eof452: ( m.cs) = 452; goto _test_eof
	_test_eof453: ( m.cs) = 453; goto _test_eof
	_test_eof454: ( m.cs) = 454; goto _test_eof
	_test_eof455: ( m.cs) = 455; goto _test_eof
	_test_eof456: ( m.cs) = 456; goto _test_eof
	_test_eof457: ( m.cs) = 457; goto _test_eof
	_test_eof458: ( m.cs) = 458; goto _test_eof
	_test_eof459: ( m.cs) = 459; goto _test_eof
	_test_eof460: ( m.cs) = 460; goto _test_eof
	_test_eof461: ( m.cs) = 461; goto _test_eof
	_test_eof462: ( m.cs) = 462; goto _test_eof
	_test_eof463: ( m.cs) = 463; goto _test_eof
	_test_eof464: ( m.cs) = 464; goto _test_eof
	_test_eof465: ( m.cs) = 465; goto _test_eof
	_test_eof466: ( m.cs) = 466; goto _test_eof
	_test_eof108: ( m.cs) = 108; goto _test_eof
	_test_eof109: ( m.cs) = 109; goto _test_eof
	_test_eof110: ( m.cs) = 110; goto _test_eof
	_test_eof111: ( m.cs) = 111; goto _test_eof
	_test_eof112: ( m.cs) = 112; goto _test_eof
	_test_eof467: ( m.cs) = 467; goto _test_eof
	_test_eof113: ( m.cs) = 113; goto _test_eof
	_test_eof468: ( m.cs) = 468; goto _test_eof
	_test_eof469: ( m.cs) = 469; goto _test_eof
	_test_eof114: ( m.cs) = 114; goto _test_eof
	_test_eof470: ( m.cs) = 470; goto _test_eof
	_test_eof471: ( m.cs) = 471; goto _test_eof
	_test_eof472: ( m.cs) = 472; goto _test_eof
	_test_eof473: ( m.cs) = 473; goto _test_eof
	_test_eof474: ( m.cs) = 474; goto _test_eof
	_test_eof475: ( m.cs) = 475; goto _test_eof
	_test_eof476: ( m.cs) = 476; goto _test_eof
	_test_eof477: ( m.cs) = 477; goto _test_eof
	_test_eof478: ( m.cs) = 478; goto _test_eof
	_test_eof115: ( m.cs) = 115; goto _test_eof
	_test_eof116: ( m.cs) = 116; goto _test_eof
	_test_eof117: ( m.cs) = 117; goto _test_eof
	_test_eof479: ( m.cs) = 479; goto _test_eof
	_test_eof118: ( m.cs) = 118; goto _test_eof
	_test_eof119: ( m.cs) = 119; goto _test_eof
	_test_eof120: ( m.cs) = 120; goto _test_eof
	_test_eof480: ( m.cs) = 480; goto _test_eof
	_test_eof121: ( m.cs) = 121; goto _test_eof
	_test_eof122: ( m.cs) = 122; goto _test_eof
	_test_eof481: ( m.cs) = 481; goto _test_eof
	_test_eof482: ( m.cs) = 482; goto _test_eof
	_test_eof123: ( m.cs) = 123; goto _test_eof
	_test_eof124: ( m.cs) = 124; goto _test_eof
	_test_eof125: ( m.cs) = 125; goto _test_eof
	_test_eof126: ( m.cs) = 126; goto _test_eof
	_test_eof483: ( m.cs) = 483; goto _test_eof
	_test_eof484: ( m.cs) = 484; goto _test_eof
	_test_eof485: ( m.cs) = 485; goto _test_eof
	_test_eof127: ( m.cs) = 127; goto _test_eof
	_test_eof486: ( m.cs) = 486; goto _test_eof
	_test_eof487: ( m.cs) = 487; goto _test_eof
	_test_eof488: ( m.cs) = 488; goto _test_eof
	_test_eof489: ( m.cs) = 489; goto _test_eof
	_test_eof490: ( m.cs) = 490; goto _test_eof
	_test_eof491: ( m.cs) = 491; goto _test_eof
	_test_eof492: ( m.cs) = 492; goto _test_eof
	_test_eof493: ( m.cs) = 493; goto _test_eof
	_test_eof494: ( m.cs) = 494; goto _test_eof
	_test_eof495: ( m.cs) = 495; goto _test_eof
	_test_eof496: ( m.cs) = 496; goto _test_eof
	_test_eof497: ( m.cs) = 497; goto _test_eof
	_test_eof498: ( m.cs) = 498; goto _test_eof
	_test_eof499: ( m.cs) = 499; goto _test_eof
	_test_eof500: ( m.cs) = 500; goto _test_eof
	_test_eof501: ( m.cs) = 501; goto _test_eof
	_test_eof502: ( m.cs) = 502; goto _test_eof
	_test_eof503: ( m.cs) = 503; goto _test_eof
	_test_eof504: ( m.cs) = 504; goto _test_eof
	_test_eof505: ( m.cs) = 505; goto _test_eof
	_test_eof128: ( m.cs) = 128; goto _test_eof
	_test_eof129: ( m.cs) = 129; goto _test_eof
	_test_eof506: ( m.cs) = 506; goto _test_eof
	_test_eof507: ( m.cs) = 507; goto _test_eof
	_test_eof508: ( m.cs) = 508; goto _test_eof
	_test_eof509: ( m.cs) = 509; goto _test_eof
	_test_eof510: ( m.cs) = 510; goto _test_eof
	_test_eof511: ( m.cs) = 511; goto _test_eof
	_test_eof512: ( m.cs) = 512; goto _test_eof
	_test_eof513: ( m.cs) = 513; goto _test_eof
	_test_eof514: ( m.cs) = 514; goto _test_eof
	_test_eof130: ( m.cs) = 130; goto _test_eof
	_test_eof131: ( m.cs) = 131; goto _test_eof
	_test_eof132: ( m.cs) = 132; goto _test_eof
	_test_eof515: ( m.cs) = 515; goto _test_eof
	_test_eof133: ( m.cs) = 133; goto _test_eof
	_test_eof134: ( m.cs) = 134; goto _test_eof
	_test_eof135: ( m.cs) = 135; goto _test_eof
	_test_eof516: ( m.cs) = 516; goto _test_eof
	_test_eof136: ( m.cs) = 136; goto _test_eof
	_test_eof137: ( m.cs) = 137; goto _test_eof
	_test_eof517: ( m.cs) = 517; goto _test_eof
	_test_eof518: ( m.cs) = 518; goto _test_eof
	_test_eof138: ( m.cs) = 138; goto _test_eof
	_test_eof139: ( m.cs) = 139; goto _test_eof
	_test_eof140: ( m.cs) = 140; goto _test_eof
	_test_eof519: ( m.cs) = 519; goto _test_eof
	_test_eof520: ( m.cs) = 520; goto _test_eof
	_test_eof141: ( m.cs) = 141; goto _test_eof
	_test_eof521: ( m.cs) = 521; goto _test_eof
	_test_eof142: ( m.cs) = 142; goto _test_eof
	_test_eof522: ( m.cs) = 522; goto _test_eof
	_test_eof523: ( m.cs) = 523; goto _test_eof
	_test_eof524: ( m.cs) = 524; goto _test_eof
	_test_eof525: ( m.cs) = 525; goto _test_eof
	_test_eof526: ( m.cs) = 526; goto _test_eof
	_test_eof527: ( m.cs) = 527; goto _test_eof
	_test_eof528: ( m.cs) = 528; goto _test_eof
	_test_eof529: ( m.cs) = 529; goto _test_eof
	_test_eof143: ( m.cs) = 143; goto _test_eof
	_test_eof144: ( m.cs) = 144; goto _test_eof
	_test_eof145: ( m.cs) = 145; goto _test_eof
	_test_eof530: ( m.cs) = 530; goto _test_eof
	_test_eof146: ( m.cs) = 146; goto _test_eof
	_test_eof147: ( m.cs) = 147; goto _test_eof
	_test_eof148: ( m.cs) = 148; goto _test_eof
	_test_eof531: ( m.cs) = 531; goto _test_eof
	_test_eof149: ( m.cs) = 149; goto _test_eof
	_test_eof150: ( m.cs) = 150; goto _test_eof
	_test_eof532: ( m.cs) = 532; goto _test_eof
	_test_eof533: ( m.cs) = 533; goto _test_eof
	_test_eof534: ( m.cs) = 534; goto _test_eof
	_test_eof535: ( m.cs) = 535; goto _test_eof
	_test_eof536: ( m.cs) = 536; goto _test_eof
	_test_eof537: ( m.cs) = 537; goto _test_eof
	_test_eof538: ( m.cs) = 538; goto _test_eof
	_test_eof539: ( m.cs) = 539; goto _test_eof
	_test_eof540: ( m.cs) = 540; goto _test_eof
	_test_eof541: ( m.cs) = 541; goto _test_eof
	_test_eof542: ( m.cs) = 542; goto _test_eof
	_test_eof543: ( m.cs) = 543; goto _test_eof
	_test_eof544: ( m.cs) = 544; goto _test_eof
	_test_eof545: ( m.cs) = 545; goto _test_eof
	_test_eof546: ( m.cs) = 546; goto _test_eof
	_test_eof547: ( m.cs) = 547; goto _test_eof
	_test_eof548: ( m.cs) = 548; goto _test_eof
	_test_eof549: ( m.cs) = 549; goto _test_eof
	_test_eof550: ( m.cs) = 550; goto _test_eof
	_test_eof551: ( m.cs) = 551; goto _test_eof
	_test_eof151: ( m.cs) = 151; goto _test_eof
	_test_eof152: ( m.cs) = 152; goto _test_eof
	_test_eof552: ( m.cs) = 552; goto _test_eof
	_test_eof553: ( m.cs) = 553; goto _test_eof
	_test_eof554: ( m.cs) = 554; goto _test_eof
	_test_eof153: ( m.cs) = 153; goto _test_eof
	_test_eof555: ( m.cs) = 555; goto _test_eof
	_test_eof556: ( m.cs) = 556; goto _test_eof
	_test_eof154: ( m.cs) = 154; goto _test_eof
	_test_eof557: ( m.cs) = 557; goto _test_eof
	_test_eof558: ( m.cs) = 558; goto _test_eof
	_test_eof559: ( m.cs) = 559; goto _test_eof
	_test_eof560: ( m.cs) = 560; goto _test_eof
	_test_eof561: ( m.cs) = 561; goto _test_eof
	_test_eof562: ( m.cs) = 562; goto _test_eof
	_test_eof563: ( m.cs) = 563; goto _test_eof
	_test_eof564: ( m.cs) = 564; goto _test_eof
	_test_eof565: ( m.cs) = 565; goto _test_eof
	_test_eof566: ( m.cs) = 566; goto _test_eof
	_test_eof567: ( m.cs) = 567; goto _test_eof
	_test_eof568: ( m.cs) = 568; goto _test_eof
	_test_eof569: ( m.cs) = 569; goto _test_eof
	_test_eof570: ( m.cs) = 570; goto _test_eof
	_test_eof571: ( m.cs) = 571; goto _test_eof
	_test_eof572: ( m.cs) = 572; goto _test_eof
	_test_eof573: ( m.cs) = 573; goto _test_eof
	_test_eof574: ( m.cs) = 574; goto _test_eof
	_test_eof155: ( m.cs) = 155; goto _test_eof
	_test_eof156: ( m.cs) = 156; goto _test_eof
	_test_eof575: ( m.cs) = 575; goto _test_eof
	_test_eof157: ( m.cs) = 157; goto _test_eof
	_test_eof576: ( m.cs) = 576; goto _test_eof
	_test_eof577: ( m.cs) = 577; goto _test_eof
	_test_eof578: ( m.cs) = 578; goto _test_eof
	_test_eof579: ( m.cs) = 579; goto _test_eof
	_test_eof580: ( m.cs) = 580; goto _test_eof
	_test_eof581: ( m.cs) = 581; goto _test_eof
	_test_eof582: ( m.cs) = 582; goto _test_eof
	_test_eof583: ( m.cs) = 583; goto _test_eof
	_test_eof158: ( m.cs) = 158; goto _test_eof
	_test_eof159: ( m.cs) = 159; goto _test_eof
	_test_eof160: ( m.cs) = 160; goto _test_eof
	_test_eof584: ( m.cs) = 584; goto _test_eof
	_test_eof161: ( m.cs) = 161; goto _test_eof
	_test_eof162: ( m.cs) = 162; goto _test_eof
	_test_eof163: ( m.cs) = 163; goto _test_eof
	_test_eof585: ( m.cs) = 585; goto _test_eof
	_test_eof164: ( m.cs) = 164; goto _test_eof
	_test_eof165: ( m.cs) = 165; goto _test_eof
	_test_eof586: ( m.cs) = 586; goto _test_eof
	_test_eof587: ( m.cs) = 587; goto _test_eof
	_test_eof166: ( m.cs) = 166; goto _test_eof
	_test_eof167: ( m.cs) = 167; goto _test_eof
	_test_eof168: ( m.cs) = 168; goto _test_eof
	_test_eof169: ( m.cs) = 169; goto _test_eof
	_test_eof170: ( m.cs) = 170; goto _test_eof
	_test_eof171: ( m.cs) = 171; goto _test_eof
	_test_eof588: ( m.cs) = 588; goto _test_eof
	_test_eof589: ( m.cs) = 589; goto _test_eof
	_test_eof590: ( m.cs) = 590; goto _test_eof
	_test_eof591: ( m.cs) = 591; goto _test_eof
	_test_eof592: ( m.cs) = 592; goto _test_eof
	_test_eof593: ( m.cs) = 593; goto _test_eof
	_test_eof594: ( m.cs) = 594; goto _test_eof
	_test_eof595: ( m.cs) = 595; goto _test_eof
	_test_eof596: ( m.cs) = 596; goto _test_eof
	_test_eof597: ( m.cs) = 597; goto _test_eof
	_test_eof598: ( m.cs) = 598; goto _test_eof
	_test_eof599: ( m.cs) = 599; goto _test_eof
	_test_eof600: ( m.cs) = 600; goto _test_eof
	_test_eof601: ( m.cs) = 601; goto _test_eof
	_test_eof602: ( m.cs) = 602; goto _test_eof
	_test_eof603: ( m.cs) = 603; goto _test_eof
	_test_eof604: ( m.cs) = 604; goto _test_eof
	_test_eof605: ( m.cs) = 605; goto _test_eof
	_test_eof606: ( m.cs) = 606; goto _test_eof
	_test_eof172: ( m.cs) = 172; goto _test_eof
	_test_eof173: ( m.cs) = 173; goto _test_eof
	_test_eof174: ( m.cs) = 174; goto _test_eof
	_test_eof607: ( m.cs) = 607; goto _test_eof
	_test_eof608: ( m.cs) = 608; goto _test_eof
	_test_eof609: ( m.cs) = 609; goto _test_eof
	_test_eof175: ( m.cs) = 175; goto _test_eof
	_test_eof610: ( m.cs) = 610; goto _test_eof
	_test_eof611: ( m.cs) = 611; goto _test_eof
	_test_eof176: ( m.cs) = 176; goto _test_eof
	_test_eof612: ( m.cs) = 612; goto _test_eof
	_test_eof613: ( m.cs) = 613; goto _test_eof
	_test_eof614: ( m.cs) = 614; goto _test_eof
	_test_eof615: ( m.cs) = 615; goto _test_eof
	_test_eof616: ( m.cs) = 616; goto _test_eof
	_test_eof177: ( m.cs) = 177; goto _test_eof
	_test_eof178: ( m.cs) = 178; goto _test_eof
	_test_eof179: ( m.cs) = 179; goto _test_eof
	_test_eof617: ( m.cs) = 617; goto _test_eof
	_test_eof180: ( m.cs) = 180; goto _test_eof
	_test_eof181: ( m.cs) = 181; goto _test_eof
	_test_eof182: ( m.cs) = 182; goto _test_eof
	_test_eof618: ( m.cs) = 618; goto _test_eof
	_test_eof183: ( m.cs) = 183; goto _test_eof
	_test_eof184: ( m.cs) = 184; goto _test_eof
	_test_eof619: ( m.cs) = 619; goto _test_eof
	_test_eof620: ( m.cs) = 620; goto _test_eof
	_test_eof185: ( m.cs) = 185; goto _test_eof
	_test_eof621: ( m.cs) = 621; goto _test_eof
	_test_eof622: ( m.cs) = 622; goto _test_eof
	_test_eof186: ( m.cs) = 186; goto _test_eof
	_test_eof187: ( m.cs) = 187; goto _test_eof
	_test_eof188: ( m.cs) = 188; goto _test_eof
	_test_eof623: ( m.cs) = 623; goto _test_eof
	_test_eof189: ( m.cs) = 189; goto _test_eof
	_test_eof190: ( m.cs) = 190; goto _test_eof
	_test_eof624: ( m.cs) = 624; goto _test_eof
	_test_eof625: ( m.cs) = 625; goto _test_eof
	_test_eof626: ( m.cs) = 626; goto _test_eof
	_test_eof627: ( m.cs) = 627; goto _test_eof
	_test_eof628: ( m.cs) = 628; goto _test_eof
	_test_eof629: ( m.cs) = 629; goto _test_eof
	_test_eof630: ( m.cs) = 630; goto _test_eof
	_test_eof631: ( m.cs) = 631; goto _test_eof
	_test_eof191: ( m.cs) = 191; goto _test_eof
	_test_eof192: ( m.cs) = 192; goto _test_eof
	_test_eof193: ( m.cs) = 193; goto _test_eof
	_test_eof632: ( m.cs) = 632; goto _test_eof
	_test_eof194: ( m.cs) = 194; goto _test_eof
	_test_eof195: ( m.cs) = 195; goto _test_eof
	_test_eof196: ( m.cs) = 196; goto _test_eof
	_test_eof633: ( m.cs) = 633; goto _test_eof
	_test_eof197: ( m.cs) = 197; goto _test_eof
	_test_eof198: ( m.cs) = 198; goto _test_eof
	_test_eof634: ( m.cs) = 634; goto _test_eof
	_test_eof635: ( m.cs) = 635; goto _test_eof
	_test_eof199: ( m.cs) = 199; goto _test_eof
	_test_eof200: ( m.cs) = 200; goto _test_eof
	_test_eof201: ( m.cs) = 201; goto _test_eof
	_test_eof636: ( m.cs) = 636; goto _test_eof
	_test_eof637: ( m.cs) = 637; goto _test_eof
	_test_eof638: ( m.cs) = 638; goto _test_eof
	_test_eof639: ( m.cs) = 639; goto _test_eof
	_test_eof640: ( m.cs) = 640; goto _test_eof
	_test_eof641: ( m.cs) = 641; goto _test_eof
	_test_eof642: ( m.cs) = 642; goto _test_eof
	_test_eof643: ( m.cs) = 643; goto _test_eof
	_test_eof644: ( m.cs) = 644; goto _test_eof
	_test_eof645: ( m.cs) = 645; goto _test_eof
	_test_eof646: ( m.cs) = 646; goto _test_eof
	_test_eof647: ( m.cs) = 647; goto _test_eof
	_test_eof648: ( m.cs) = 648; goto _test_eof
	_test_eof649: ( m.cs) = 649; goto _test_eof
	_test_eof650: ( m.cs) = 650; goto _test_eof
	_test_eof651: ( m.cs) = 651; goto _test_eof
	_test_eof652: ( m.cs) = 652; goto _test_eof
	_test_eof653: ( m.cs) = 653; goto _test_eof
	_test_eof654: ( m.cs) = 654; goto _test_eof
	_test_eof202: ( m.cs) = 202; goto _test_eof
	_test_eof203: ( m.cs) = 203; goto _test_eof
	_test_eof204: ( m.cs) = 204; goto _test_eof
	_test_eof205: ( m.cs) = 205; goto _test_eof
	_test_eof206: ( m.cs) = 206; goto _test_eof
	_test_eof655: ( m.cs) = 655; goto _test_eof
	_test_eof207: ( m.cs) = 207; goto _test_eof
	_test_eof208: ( m.cs) = 208; goto _test_eof
	_test_eof656: ( m.cs) = 656; goto _test_eof
	_test_eof657: ( m.cs) = 657; goto _test_eof
	_test_eof658: ( m.cs) = 658; goto _test_eof
	_test_eof659: ( m.cs) = 659; goto _test_eof
	_test_eof660: ( m.cs) = 660; goto _test_eof
	_test_eof661: ( m.cs) = 661; goto _test_eof
	_test_eof662: ( m.cs) = 662; goto _test_eof
	_test_eof663: ( m.cs) = 663; goto _test_eof
	_test_eof664: ( m.cs) = 664; goto _test_eof
	_test_eof209: ( m.cs) = 209; goto _test_eof
	_test_eof210: ( m.cs) = 210; goto _test_eof
	_test_eof211: ( m.cs) = 211; goto _test_eof
	_test_eof665: ( m.cs) = 665; goto _test_eof
	_test_eof212: ( m.cs) = 212; goto _test_eof
	_test_eof213: ( m.cs) = 213; goto _test_eof
	_test_eof214: ( m.cs) = 214; goto _test_eof
	_test_eof666: ( m.cs) = 666; goto _test_eof
	_test_eof215: ( m.cs) = 215; goto _test_eof
	_test_eof216: ( m.cs) = 216; goto _test_eof
	_test_eof667: ( m.cs) = 667; goto _test_eof
	_test_eof668: ( m.cs) = 668; goto _test_eof
	_test_eof217: ( m.cs) = 217; goto _test_eof
	_test_eof218: ( m.cs) = 218; goto _test_eof
	_test_eof219: ( m.cs) = 219; goto _test_eof
	_test_eof220: ( m.cs) = 220; goto _test_eof
	_test_eof669: ( m.cs) = 669; goto _test_eof
	_test_eof221: ( m.cs) = 221; goto _test_eof
	_test_eof222: ( m.cs) = 222; goto _test_eof
	_test_eof670: ( m.cs) = 670; goto _test_eof
	_test_eof671: ( m.cs) = 671; goto _test_eof
	_test_eof672: ( m.cs) = 672; goto _test_eof
	_test_eof673: ( m.cs) = 673; goto _test_eof
	_test_eof674: ( m.cs) = 674; goto _test_eof
	_test_eof675: ( m.cs) = 675; goto _test_eof
	_test_eof676: ( m.cs) = 676; goto _test_eof
	_test_eof677: ( m.cs) = 677; goto _test_eof
	_test_eof223: ( m.cs) = 223; goto _test_eof
	_test_eof224: ( m.cs) = 224; goto _test_eof
	_test_eof225: ( m.cs) = 225; goto _test_eof
	_test_eof678: ( m.cs) = 678; goto _test_eof
	_test_eof226: ( m.cs) = 226; goto _test_eof
	_test_eof227: ( m.cs) = 227; goto _test_eof
	_test_eof228: ( m.cs) = 228; goto _test_eof
	_test_eof679: ( m.cs) = 679; goto _test_eof
	_test_eof229: ( m.cs) = 229; goto _test_eof
	_test_eof230: ( m.cs) = 230; goto _test_eof
	_test_eof680: ( m.cs) = 680; goto _test_eof
	_test_eof681: ( m.cs) = 681; goto _test_eof
	_test_eof231: ( m.cs) = 231; goto _test_eof
	_test_eof232: ( m.cs) = 232; goto _test_eof
	_test_eof233: ( m.cs) = 233; goto _test_eof
	_test_eof682: ( m.cs) = 682; goto _test_eof
	_test_eof683: ( m.cs) = 683; goto _test_eof
	_test_eof684: ( m.cs) = 684; goto _test_eof
	_test_eof685: ( m.cs) = 685; goto _test_eof
	_test_eof686: ( m.cs) = 686; goto _test_eof
	_test_eof687: ( m.cs) = 687; goto _test_eof
	_test_eof688: ( m.cs) = 688; goto _test_eof
	_test_eof689: ( m.cs) = 689; goto _test_eof
	_test_eof690: ( m.cs) = 690; goto _test_eof
	_test_eof691: ( m.cs) = 691; goto _test_eof
	_test_eof692: ( m.cs) = 692; goto _test_eof
	_test_eof693: ( m.cs) = 693; goto _test_eof
	_test_eof694: ( m.cs) = 694; goto _test_eof
	_test_eof695: ( m.cs) = 695; goto _test_eof
	_test_eof696: ( m.cs) = 696; goto _test_eof
	_test_eof697: ( m.cs) = 697; goto _test_eof
	_test_eof698: ( m.cs) = 698; goto _test_eof
	_test_eof699: ( m.cs) = 699; goto _test_eof
	_test_eof700: ( m.cs) = 700; goto _test_eof
	_test_eof234: ( m.cs) = 234; goto _test_eof
	_test_eof235: ( m.cs) = 235; goto _test_eof
	_test_eof701: ( m.cs) = 701; goto _test_eof
	_test_eof236: ( m.cs) = 236; goto _test_eof
	_test_eof237: ( m.cs) = 237; goto _test_eof
	_test_eof702: ( m.cs) = 702; goto _test_eof
	_test_eof703: ( m.cs) = 703; goto _test_eof
	_test_eof704: ( m.cs) = 704; goto _test_eof
	_test_eof705: ( m.cs) = 705; goto _test_eof
	_test_eof706: ( m.cs) = 706; goto _test_eof
	_test_eof707: ( m.cs) = 707; goto _test_eof
	_test_eof708: ( m.cs) = 708; goto _test_eof
	_test_eof709: ( m.cs) = 709; goto _test_eof
	_test_eof238: ( m.cs) = 238; goto _test_eof
	_test_eof239: ( m.cs) = 239; goto _test_eof
	_test_eof240: ( m.cs) = 240; goto _test_eof
	_test_eof710: ( m.cs) = 710; goto _test_eof
	_test_eof241: ( m.cs) = 241; goto _test_eof
	_test_eof242: ( m.cs) = 242; goto _test_eof
	_test_eof243: ( m.cs) = 243; goto _test_eof
	_test_eof711: ( m.cs) = 711; goto _test_eof
	_test_eof244: ( m.cs) = 244; goto _test_eof
	_test_eof245: ( m.cs) = 245; goto _test_eof
	_test_eof712: ( m.cs) = 712; goto _test_eof
	_test_eof713: ( m.cs) = 713; goto _test_eof
	_test_eof246: ( m.cs) = 246; goto _test_eof
	_test_eof247: ( m.cs) = 247; goto _test_eof
	_test_eof714: ( m.cs) = 714; goto _test_eof
	_test_eof250: ( m.cs) = 250; goto _test_eof
	_test_eof717: ( m.cs) = 717; goto _test_eof
	_test_eof718: ( m.cs) = 718; goto _test_eof
	_test_eof251: ( m.cs) = 251; goto _test_eof
	_test_eof252: ( m.cs) = 252; goto _test_eof
	_test_eof253: ( m.cs) = 253; goto _test_eof
	_test_eof254: ( m.cs) = 254; goto _test_eof
	_test_eof719: ( m.cs) = 719; goto _test_eof
	_test_eof255: ( m.cs) = 255; goto _test_eof
	_test_eof720: ( m.cs) = 720; goto _test_eof
	_test_eof256: ( m.cs) = 256; goto _test_eof
	_test_eof257: ( m.cs) = 257; goto _test_eof
	_test_eof258: ( m.cs) = 258; goto _test_eof
	_test_eof715: ( m.cs) = 715; goto _test_eof
	_test_eof716: ( m.cs) = 716; goto _test_eof
	_test_eof248: ( m.cs) = 248; goto _test_eof
	_test_eof249: ( m.cs) = 249; goto _test_eof

	_test_eof: {}
	if ( m.p) == ( m.eof) {
		switch ( m.cs) {
		case 9, 250:
//line plugins/parsers/influx/machine.go.rl:23

	err = ErrNameParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

		case 2, 3, 4, 5, 6, 7, 8, 29, 32, 33, 36, 37, 38, 50, 51, 52, 53, 54, 74, 76, 77, 94, 104, 106, 142, 154, 157, 172, 173, 174, 175, 176, 177, 178, 179, 180, 181, 182, 183, 184, 185, 186, 202, 203, 204, 205, 206, 207, 208, 209, 210, 211, 212, 213, 214, 215, 216, 217, 218, 234, 235, 236, 237, 238, 239, 240, 241, 242, 243, 244, 245, 246:
//line plugins/parsers/influx/machine.go.rl:30

	err = ErrFieldParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

		case 14, 15, 16, 23, 25, 26, 252, 253, 254, 255, 256, 257:
//line plugins/parsers/influx/machine.go.rl:37

	err = ErrTagParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

		case 233:
//line plugins/parsers/influx/machine.go.rl:44

	err = ErrTimestampParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

		case 259:
//line plugins/parsers/influx/machine.go.rl:73

	foundMetric = true

		case 289, 292, 296, 364, 388, 389, 393, 394, 395, 519, 553, 554, 556, 717:
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

		case 340, 341, 342, 344, 363, 419, 443, 444, 448, 468, 484, 485, 487, 719, 720:
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

		case 613, 659, 704:
//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

		case 614, 662, 707:
//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

		case 315, 608, 609, 611, 612, 615, 621, 622, 655, 656, 657, 658, 660, 661, 663, 701, 702, 703, 705, 706, 708:
//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

		case 616, 617, 618, 619, 620, 664, 665, 666, 667, 668, 709, 710, 711, 712, 713:
//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

		case 265, 268, 269, 270, 271, 272, 273, 274, 275, 276, 277, 278, 279, 280, 281, 282, 283, 284, 285, 320, 322, 323, 324, 325, 326, 327, 328, 329, 330, 331, 332, 333, 334, 335, 336, 337, 338, 339, 367, 370, 371, 372, 373, 374, 375, 376, 377, 378, 379, 380, 381, 382, 383, 384, 385, 386, 387, 399, 401, 402, 403, 404, 405, 406, 407, 408, 409, 410, 411, 412, 413, 414, 415, 416, 417, 418, 422, 425, 426, 427, 428, 429, 430, 431, 432, 433, 434, 435, 436, 437, 438, 439, 440, 441, 442, 588, 589, 590, 591, 592, 593, 594, 595, 596, 597, 598, 599, 600, 601, 602, 603, 604, 605, 606, 636, 637, 638, 639, 640, 641, 642, 643, 644, 645, 646, 647, 648, 649, 650, 651, 652, 653, 654, 682, 683, 684, 685, 686, 687, 688, 689, 690, 691, 692, 693, 694, 695, 696, 697, 698, 699, 700:
//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

		case 11, 39, 41, 166, 168:
//line plugins/parsers/influx/machine.go.rl:23

	err = ErrNameParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:30

	err = ErrFieldParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

		case 35, 75, 105, 171, 201:
//line plugins/parsers/influx/machine.go.rl:30

	err = ErrFieldParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:44

	err = ErrTimestampParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

		case 21, 45, 46, 47, 59, 60, 62, 64, 69, 71, 72, 78, 79, 80, 85, 87, 89, 90, 98, 99, 101, 102, 103, 108, 109, 110, 123, 124, 138, 139:
//line plugins/parsers/influx/machine.go.rl:37

	err = ErrTagParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:30

	err = ErrFieldParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

		case 61:
//line plugins/parsers/influx/machine.go.rl:37

	err = ErrTagParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:44

	err = ErrTimestampParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

		case 1:
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:37

	err = ErrTagParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

		case 524, 578, 672:
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

		case 527, 581, 675:
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

		case 396, 520, 521, 522, 523, 525, 526, 528, 552, 575, 576, 577, 579, 580, 582, 669, 670, 671, 673, 674, 676:
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

		case 529, 530, 531, 532, 533, 583, 584, 585, 586, 587, 677, 678, 679, 680, 681:
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

		case 293, 297, 298, 299, 300, 301, 302, 303, 304, 305, 306, 307, 308, 309, 310, 311, 312, 313, 314, 390, 534, 535, 536, 537, 538, 539, 540, 541, 542, 543, 544, 545, 546, 547, 548, 549, 550, 551, 555, 557, 558, 559, 560, 561, 562, 563, 564, 565, 566, 567, 568, 569, 570, 571, 572, 573, 574:
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

		case 17, 24:
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:37

	err = ErrTagParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

		case 473, 509, 626:
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:103

	err = m.handler.AddInt(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

		case 476, 512, 629:
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:112

	err = m.handler.AddUint(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

		case 467, 469, 470, 471, 472, 474, 475, 477, 483, 506, 507, 508, 510, 511, 513, 623, 624, 625, 627, 628, 630:
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:121

	err = m.handler.AddFloat(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

		case 478, 479, 480, 481, 482, 514, 515, 516, 517, 518, 631, 632, 633, 634, 635:
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:130

	err = m.handler.AddBool(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

		case 343, 345, 346, 347, 348, 349, 350, 351, 352, 353, 354, 355, 356, 357, 358, 359, 360, 361, 362, 445, 449, 450, 451, 452, 453, 454, 455, 456, 457, 458, 459, 460, 461, 462, 463, 464, 465, 466, 486, 488, 489, 490, 491, 492, 493, 494, 495, 496, 497, 498, 499, 500, 501, 502, 503, 504, 505:
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:148

	err = m.handler.SetTimestamp(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

		case 10:
//line plugins/parsers/influx/machine.go.rl:23

	err = ErrNameParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:37

	err = ErrTagParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

		case 100:
//line plugins/parsers/influx/machine.go.rl:37

	err = ErrTagParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:30

	err = ErrFieldParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:44

	err = ErrTimestampParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

		case 12, 13, 27, 28, 30, 31, 42, 43, 55, 56, 57, 58, 73, 92, 93, 95, 97, 140, 141, 143, 144, 145, 146, 147, 148, 149, 150, 151, 152, 155, 156, 158, 159, 160, 161, 162, 163, 164, 165, 219, 220, 221, 222, 223, 224, 225, 226, 227, 228, 229, 230:
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:37

	err = ErrTagParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:30

	err = ErrFieldParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

		case 18, 19, 20, 22, 48, 49, 65, 66, 67, 68, 70, 81, 82, 83, 84, 86, 88, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 125, 126, 128, 129, 130, 131, 132, 133, 134, 135, 136, 137, 187, 188, 189, 190, 191, 192, 193, 194, 195, 196, 197, 198:
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:37

	err = ErrTagParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:30

	err = ErrFieldParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

		case 40, 167, 169, 170, 199, 200, 231, 232:
//line plugins/parsers/influx/machine.go.rl:23

	err = ErrNameParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:37

	err = ErrTagParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:30

	err = ErrFieldParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

		case 44, 91, 153:
//line plugins/parsers/influx/machine.go.rl:77

	err = m.handler.SetMeasurement(m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:37

	err = ErrTagParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:30

	err = ErrFieldParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:44

	err = ErrTimestampParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

		case 63, 107, 127:
//line plugins/parsers/influx/machine.go.rl:90

	err = m.handler.AddTag(key, m.text())
	if err != nil {
		( m.p)--

		( m.cs) = 247;
		{( m.p)++; ( m.cs) = 0; goto _out }
	}

//line plugins/parsers/influx/machine.go.rl:37

	err = ErrTagParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:30

	err = ErrFieldParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

//line plugins/parsers/influx/machine.go.rl:44

	err = ErrTimestampParse
	( m.p)--

	( m.cs) = 247;
	{( m.p)++; ( m.cs) = 0; goto _out }

//line plugins/parsers/influx/machine.go:30741
		}
	}

	_out: {}
	}

//line plugins/parsers/influx/machine.go.rl:390

	if err != nil {
		return err
	}

	// This would indicate an error in the machine that was reported with a
	// more specific error.  We return a generic error but this should
	// possibly be a panic.
	if m.cs == 0 {
		m.cs = LineProtocol_en_discard_line
		return ErrParse
	}

	// If we haven't found a metric line yet and we reached the EOF, report it
	// now.  This happens when the data ends with a comment or whitespace.
	//
	// Otherwise we have successfully parsed a metric line, so if we are at
	// the EOF we will report it the next call.
	if !foundMetric && m.p == m.pe && m.pe == m.eof {
		return EOF
	}

	return nil
}

// Position returns the current byte offset into the data.
func (m *machine) Position() int {
	return m.p
}

// LineOffset returns the byte offset of the current line.
func (m *machine) LineOffset() int {
	return m.sol
}

// LineNumber returns the current line number.  Lines are counted based on the
// regular expression `\r?\n`.
func (m *machine) LineNumber() int {
	return m.lineno
}

// Column returns the current column.
func (m *machine) Column() int {
	lineOffset := m.p - m.sol
	return lineOffset + 1
}

func (m *machine) text() []byte {
	return m.data[m.pb:m.p]
}
