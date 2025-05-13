/*
 * Copyright (C) 2017-2022 Alexey Khokholov (Nuke.YKT)
 *
 * This file is part of Nuked OPN2.
 *
 * This library is free software; you can redistribute it and/or
 * modify it under the terms of the GNU Lesser General Public
 * License as published by the Free Software Foundation; either
 * version 2.1 of the License, or (at your option) any later version.
 *
 * This library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public
 * License along with this library; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 *
 *  Nuked OPN2(Yamaha YM3438) emulator.
 *  Thanks:
 *      Silicon Pr0n:
 *          Yamaha YM3438 decap and die shot(digshadow).
 *      OPLx decapsulated(Matthew Gambrell, Olli Niemitalo):
 *          OPL2 ROMs.
 *
 * version: 1.0.12
 */
package nukeykt

import (
	"github.com/elemir/cbool"
)

const (
	OUTPUT_FACTOR   = 11
	OUTPUT_FACTOR_F = 12
	FILTER_CUTOFF   = 0.512331301282628 // 5894Hz  single pole IIR low pass
	FILTER_CUTOFF_I = (1 - FILTER_CUTOFF)

	RSM_FRAC           = 10
	OPN_WRITEBUF_SIZE  = 2048
	OPN_WRITEBUF_DELAY = 15

	ModeYM2612   = 0x01 /* Enables YM2612 emulation (MD1, MD2 VA2) */
	ModeReadmode = 0x02 /* Enables status read on any port (TeraDrive, MD1 VA7, MD2, etc) */
)

type YM3438 struct {
	cycles   uint32
	channel  uint32
	mol, mor int16
	/* IO */
	write_data       uint16
	write_a          uint8
	write_d          uint8
	write_a_en       uint8
	write_d_en       uint8
	write_busy       uint8
	write_busy_cnt   uint8
	write_fm_address uint8
	write_fm_data    uint8
	write_fm_mode_a  uint16
	address          uint16
	data             uint8
	pin_test_in      uint8
	pin_irq          uint8
	busy             uint8
	/* LFO */
	lfo_en       uint8
	lfo_freq     uint8
	lfo_pm       uint8
	lfo_am       uint8
	lfo_cnt      uint8
	lfo_inc      uint8
	lfo_quotient uint8
	/* Phase generator */
	pg_fnum  uint16
	pg_block uint8
	pg_kcode uint8
	pg_inc   [24]uint32
	pg_phase [24]uint32
	pg_reset [24]uint8
	pg_read  uint32
	/* Envelope generator */
	eg_cycle             uint8
	eg_cycle_stop        uint8
	eg_shift             uint8
	eg_shift_lock        uint8
	eg_timer_low_lock    uint8
	eg_timer             uint16
	eg_timer_inc         uint8
	eg_quotient          uint16
	eg_custom_timer      uint8
	eg_rate              uint8
	eg_ksv               uint8
	eg_inc               uint8
	eg_ratemax           uint8
	eg_sl                [2]uint8
	eg_lfo_am            uint8
	eg_tl                [2]uint8
	eg_state             [24]uint8
	eg_level             [24]uint16
	eg_out               [24]uint16
	eg_kon               [24]uint8
	eg_kon_csm           [24]uint8
	eg_kon_latch         [24]uint8
	eg_csm_mode          [24]uint8
	eg_ssg_enable        [24]uint8
	eg_ssg_pgrst_latch   [24]uint8
	eg_ssg_repeat_latch  [24]uint8
	eg_ssg_hold_up_latch [24]uint8
	eg_ssg_dir           [24]uint8
	eg_ssg_inv           [24]uint8
	eg_read              [2]uint32
	eg_read_inc          uint8
	/* FM */
	fm_op1 [6][2]int16
	fm_op2 [6]int16
	fm_out [24]int16
	fm_mod [24]uint16
	/* Channel */
	ch_acc    [6]int16
	ch_out    [6]int16
	ch_lock   int16
	ch_lock_l uint8
	ch_lock_r uint8
	ch_read   int16
	/* Timer */
	timer_a_cnt           uint16
	timer_a_reg           uint16
	timer_a_load_lock     uint8
	timer_a_load          uint8
	timer_a_enable        uint8
	timer_a_reset         uint8
	timer_a_load_latch    uint8
	timer_a_overflow_flag uint8
	timer_a_overflow      uint8

	timer_b_cnt           uint16
	timer_b_subcnt        uint8
	timer_b_reg           uint16
	timer_b_load_lock     uint8
	timer_b_load          uint8
	timer_b_enable        uint8
	timer_b_reset         uint8
	timer_b_load_latch    uint8
	timer_b_overflow_flag uint8
	timer_b_overflow      uint8

	/* Register set */
	mode_test_21      [8]uint8
	mode_test_2c      [8]uint8
	mode_ch3          uint8
	mode_kon_channel  uint8
	mode_kon_operator [4]uint8
	mode_kon          [24]uint8
	mode_csm          uint8
	mode_kon_csm      uint8
	dacen             uint8
	dacdata           int16

	ks     [24]uint8
	ar     [24]uint8
	sr     [24]uint8
	dt     [24]uint8
	multi  [24]uint8
	sl     [24]uint8
	rr     [24]uint8
	dr     [24]uint8
	am     [24]uint8
	tl     [24]uint8
	ssg_eg [24]uint8

	fnum         [6]uint16
	block        [6]uint8
	kcode        [6]uint8
	fnum_3ch     [6]uint16
	block_3ch    [6]uint8
	kcode_3ch    [6]uint8
	reg_a4       uint8
	reg_ac       uint8
	connect      [6]uint8
	fb           [6]uint8
	pan_l, pan_r [6]uint8
	ams          [6]uint8
	pms          [6]uint8
	status       uint8
	status_time  uint32

	mute       [7]uint32
	rateratio  int32
	samplecnt  int32
	oldsamples [2]int32
	samples    [2]int32

	writebuf_samplecnt uint64
	writebuf_cur       uint32
	writebuf_last      uint32
	writebuf_lasttime  uint64
	writebuf           [OPN_WRITEBUF_SIZE]writebuf
}

type writebuf struct {
	time uint64
	port uint8
	data uint8
}

func SIGN_EXTEND(bit_index, value int16) int16 {
	return (((value) & ((1 << (bit_index)) - 1)) - ((value) & (1 << (bit_index))))
}

const (
	eg_num_attack  = 0
	eg_num_decay   = 1
	eg_num_sustain = 2
	eg_num_release = 3
)

/* logsin table */
var (
	logsinrom = [256]uint16{
		0x859, 0x6c3, 0x607, 0x58b, 0x52e, 0x4e4, 0x4a6, 0x471,
		0x443, 0x41a, 0x3f5, 0x3d3, 0x3b5, 0x398, 0x37e, 0x365,
		0x34e, 0x339, 0x324, 0x311, 0x2ff, 0x2ed, 0x2dc, 0x2cd,
		0x2bd, 0x2af, 0x2a0, 0x293, 0x286, 0x279, 0x26d, 0x261,
		0x256, 0x24b, 0x240, 0x236, 0x22c, 0x222, 0x218, 0x20f,
		0x206, 0x1fd, 0x1f5, 0x1ec, 0x1e4, 0x1dc, 0x1d4, 0x1cd,
		0x1c5, 0x1be, 0x1b7, 0x1b0, 0x1a9, 0x1a2, 0x19b, 0x195,
		0x18f, 0x188, 0x182, 0x17c, 0x177, 0x171, 0x16b, 0x166,
		0x160, 0x15b, 0x155, 0x150, 0x14b, 0x146, 0x141, 0x13c,
		0x137, 0x133, 0x12e, 0x129, 0x125, 0x121, 0x11c, 0x118,
		0x114, 0x10f, 0x10b, 0x107, 0x103, 0x0ff, 0x0fb, 0x0f8,
		0x0f4, 0x0f0, 0x0ec, 0x0e9, 0x0e5, 0x0e2, 0x0de, 0x0db,
		0x0d7, 0x0d4, 0x0d1, 0x0cd, 0x0ca, 0x0c7, 0x0c4, 0x0c1,
		0x0be, 0x0bb, 0x0b8, 0x0b5, 0x0b2, 0x0af, 0x0ac, 0x0a9,
		0x0a7, 0x0a4, 0x0a1, 0x09f, 0x09c, 0x099, 0x097, 0x094,
		0x092, 0x08f, 0x08d, 0x08a, 0x088, 0x086, 0x083, 0x081,
		0x07f, 0x07d, 0x07a, 0x078, 0x076, 0x074, 0x072, 0x070,
		0x06e, 0x06c, 0x06a, 0x068, 0x066, 0x064, 0x062, 0x060,
		0x05e, 0x05c, 0x05b, 0x059, 0x057, 0x055, 0x053, 0x052,
		0x050, 0x04e, 0x04d, 0x04b, 0x04a, 0x048, 0x046, 0x045,
		0x043, 0x042, 0x040, 0x03f, 0x03e, 0x03c, 0x03b, 0x039,
		0x038, 0x037, 0x035, 0x034, 0x033, 0x031, 0x030, 0x02f,
		0x02e, 0x02d, 0x02b, 0x02a, 0x029, 0x028, 0x027, 0x026,
		0x025, 0x024, 0x023, 0x022, 0x021, 0x020, 0x01f, 0x01e,
		0x01d, 0x01c, 0x01b, 0x01a, 0x019, 0x018, 0x017, 0x017,
		0x016, 0x015, 0x014, 0x014, 0x013, 0x012, 0x011, 0x011,
		0x010, 0x00f, 0x00f, 0x00e, 0x00d, 0x00d, 0x00c, 0x00c,
		0x00b, 0x00a, 0x00a, 0x009, 0x009, 0x008, 0x008, 0x007,
		0x007, 0x007, 0x006, 0x006, 0x005, 0x005, 0x005, 0x004,
		0x004, 0x004, 0x003, 0x003, 0x003, 0x002, 0x002, 0x002,
		0x002, 0x001, 0x001, 0x001, 0x001, 0x001, 0x001, 0x001,
		0x000, 0x000, 0x000, 0x000, 0x000, 0x000, 0x000, 0x000,
	}

	/* exp table */
	exprom = [256]uint16{
		0x000, 0x003, 0x006, 0x008, 0x00b, 0x00e, 0x011, 0x014,
		0x016, 0x019, 0x01c, 0x01f, 0x022, 0x025, 0x028, 0x02a,
		0x02d, 0x030, 0x033, 0x036, 0x039, 0x03c, 0x03f, 0x042,
		0x045, 0x048, 0x04b, 0x04e, 0x051, 0x054, 0x057, 0x05a,
		0x05d, 0x060, 0x063, 0x066, 0x069, 0x06c, 0x06f, 0x072,
		0x075, 0x078, 0x07b, 0x07e, 0x082, 0x085, 0x088, 0x08b,
		0x08e, 0x091, 0x094, 0x098, 0x09b, 0x09e, 0x0a1, 0x0a4,
		0x0a8, 0x0ab, 0x0ae, 0x0b1, 0x0b5, 0x0b8, 0x0bb, 0x0be,
		0x0c2, 0x0c5, 0x0c8, 0x0cc, 0x0cf, 0x0d2, 0x0d6, 0x0d9,
		0x0dc, 0x0e0, 0x0e3, 0x0e7, 0x0ea, 0x0ed, 0x0f1, 0x0f4,
		0x0f8, 0x0fb, 0x0ff, 0x102, 0x106, 0x109, 0x10c, 0x110,
		0x114, 0x117, 0x11b, 0x11e, 0x122, 0x125, 0x129, 0x12c,
		0x130, 0x134, 0x137, 0x13b, 0x13e, 0x142, 0x146, 0x149,
		0x14d, 0x151, 0x154, 0x158, 0x15c, 0x160, 0x163, 0x167,
		0x16b, 0x16f, 0x172, 0x176, 0x17a, 0x17e, 0x181, 0x185,
		0x189, 0x18d, 0x191, 0x195, 0x199, 0x19c, 0x1a0, 0x1a4,
		0x1a8, 0x1ac, 0x1b0, 0x1b4, 0x1b8, 0x1bc, 0x1c0, 0x1c4,
		0x1c8, 0x1cc, 0x1d0, 0x1d4, 0x1d8, 0x1dc, 0x1e0, 0x1e4,
		0x1e8, 0x1ec, 0x1f0, 0x1f5, 0x1f9, 0x1fd, 0x201, 0x205,
		0x209, 0x20e, 0x212, 0x216, 0x21a, 0x21e, 0x223, 0x227,
		0x22b, 0x230, 0x234, 0x238, 0x23c, 0x241, 0x245, 0x249,
		0x24e, 0x252, 0x257, 0x25b, 0x25f, 0x264, 0x268, 0x26d,
		0x271, 0x276, 0x27a, 0x27f, 0x283, 0x288, 0x28c, 0x291,
		0x295, 0x29a, 0x29e, 0x2a3, 0x2a8, 0x2ac, 0x2b1, 0x2b5,
		0x2ba, 0x2bf, 0x2c4, 0x2c8, 0x2cd, 0x2d2, 0x2d6, 0x2db,
		0x2e0, 0x2e5, 0x2e9, 0x2ee, 0x2f3, 0x2f8, 0x2fd, 0x302,
		0x306, 0x30b, 0x310, 0x315, 0x31a, 0x31f, 0x324, 0x329,
		0x32e, 0x333, 0x338, 0x33d, 0x342, 0x347, 0x34c, 0x351,
		0x356, 0x35b, 0x360, 0x365, 0x36a, 0x370, 0x375, 0x37a,
		0x37f, 0x384, 0x38a, 0x38f, 0x394, 0x399, 0x39f, 0x3a4,
		0x3a9, 0x3ae, 0x3b4, 0x3b9, 0x3bf, 0x3c4, 0x3c9, 0x3cf,
		0x3d4, 0x3da, 0x3df, 0x3e4, 0x3ea, 0x3ef, 0x3f5, 0x3fa,
	}

	/* Note table */
	fn_note = [16]uint32{
		0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 3, 3, 3, 3, 3, 3,
	}

	/* Envelope generator */
	eg_stephi = [4][4]uint32{
		{0, 0, 0, 0},
		{1, 0, 0, 0},
		{1, 0, 1, 0},
		{1, 1, 1, 0},
	}

	eg_am_shift = [4]uint8{
		7, 3, 1, 0,
	}

	/* Phase generator */
	pg_detune = [8]uint32{16, 17, 19, 20, 22, 24, 27, 29}

	pg_lfo_sh1 = [8][8]uint32{
		{7, 7, 7, 7, 7, 7, 7, 7},
		{7, 7, 7, 7, 7, 7, 7, 7},
		{7, 7, 7, 7, 7, 7, 1, 1},
		{7, 7, 7, 7, 1, 1, 1, 1},
		{7, 7, 7, 1, 1, 1, 1, 0},
		{7, 7, 1, 1, 0, 0, 0, 0},
		{7, 7, 1, 1, 0, 0, 0, 0},
		{7, 7, 1, 1, 0, 0, 0, 0},
	}

	pg_lfo_sh2 = [8][8]uint32{
		{7, 7, 7, 7, 7, 7, 7, 7},
		{7, 7, 7, 7, 2, 2, 2, 2},
		{7, 7, 7, 2, 2, 2, 7, 7},
		{7, 7, 2, 2, 7, 7, 2, 2},
		{7, 7, 2, 7, 7, 7, 2, 7},
		{7, 7, 7, 2, 7, 7, 2, 1},
		{7, 7, 7, 2, 7, 7, 2, 1},
		{7, 7, 7, 2, 7, 7, 2, 1},
	}

	/* Address decoder */
	op_offset = [12]uint32{
		0x000, /* Ch1 OP1/OP2 */
		0x001, /* Ch2 OP1/OP2 */
		0x002, /* Ch3 OP1/OP2 */
		0x100, /* Ch4 OP1/OP2 */
		0x101, /* Ch5 OP1/OP2 */
		0x102, /* Ch6 OP1/OP2 */
		0x004, /* Ch1 OP3/OP4 */
		0x005, /* Ch2 OP3/OP4 */
		0x006, /* Ch3 OP3/OP4 */
		0x104, /* Ch4 OP3/OP4 */
		0x105, /* Ch5 OP3/OP4 */
		0x106, /* Ch6 OP3/OP4 */
	}

	ch_offset = [6]uint32{
		0x000, /* Ch1 */
		0x001, /* Ch2 */
		0x002, /* Ch3 */
		0x100, /* Ch4 */
		0x101, /* Ch5 */
		0x102, /* Ch6 */
	}

	/* LFO */
	lfo_cycles = [8]uint32{
		108, 77, 71, 67, 62, 44, 8, 5,
	}

	/* FM algorithm */
	fm_algorithm = [4][6][8]uint32{
		{
			{1, 1, 1, 1, 1, 1, 1, 1}, /* OP1_0         */
			{1, 1, 1, 1, 1, 1, 1, 1}, /* OP1_1         */
			{0, 0, 0, 0, 0, 0, 0, 0}, /* OP2           */
			{0, 0, 0, 0, 0, 0, 0, 0}, /* Last operator */
			{0, 0, 0, 0, 0, 0, 0, 0}, /* Last operator */
			{0, 0, 0, 0, 0, 0, 0, 1}, /* Out           */
		},
		{
			{0, 1, 0, 0, 0, 1, 0, 0}, /* OP1_0         */
			{0, 0, 0, 0, 0, 0, 0, 0}, /* OP1_1         */
			{1, 1, 1, 0, 0, 0, 0, 0}, /* OP2           */
			{0, 0, 0, 0, 0, 0, 0, 0}, /* Last operator */
			{0, 0, 0, 0, 0, 0, 0, 0}, /* Last operator */
			{0, 0, 0, 0, 0, 1, 1, 1}, /* Out           */
		},
		{
			{0, 0, 0, 0, 0, 0, 0, 0}, /* OP1_0         */
			{0, 0, 0, 0, 0, 0, 0, 0}, /* OP1_1         */
			{0, 0, 0, 0, 0, 0, 0, 0}, /* OP2           */
			{1, 0, 0, 1, 1, 1, 1, 0}, /* Last operator */
			{0, 0, 0, 0, 0, 0, 0, 0}, /* Last operator */
			{0, 0, 0, 0, 1, 1, 1, 1}, /* Out           */
		},
		{
			{0, 0, 1, 0, 0, 1, 0, 0}, /* OP1_0         */
			{0, 0, 0, 0, 0, 0, 0, 0}, /* OP1_1         */
			{0, 0, 0, 1, 0, 0, 0, 0}, /* OP2           */
			{1, 1, 0, 1, 1, 0, 0, 0}, /* Last operator */
			{0, 0, 1, 0, 0, 0, 0, 0}, /* Last operator */
			{1, 1, 1, 1, 1, 1, 1, 1}, /* Out           */
		},
	}

	chip_type uint32 = ModeReadmode
)

/*
 * Copyright (C) 2017-2022 Alexey Khokholov (Nuke.YKT)
 *
 * This file is part of Nuked OPN2.
 *
 * This library is free software; you can redistribute it and/or
 * modify it under the terms of the GNU Lesser General Public
 * License as published by the Free Software Foundation; either
 * version 2.1 of the License, or (at your option) any later version.
 *
 * This library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public
 * License along with this library; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 *
 *  Nuked OPN2(Yamaha YM3438) emulator.
 *  Thanks:
 *      Silicon Pr0n:
 *          Yamaha YM3438 decap and die shot(digshadow).
 *      OPLx decapsulated(Matthew Gambrell, Olli Niemitalo):
 *          OPL2 ROMs.
 *
 * version: 1.0.12
 */

func OPN2_DoIO(chip *YM3438) {
	/* Write signal check */
	chip.write_a_en = cbool.ToInt[uint8]((chip.write_a & 0x03) == 0x01)
	chip.write_d_en = cbool.ToInt[uint8]((chip.write_d & 0x03) == 0x01)
	chip.write_a <<= 1
	chip.write_d <<= 1
	/* Busy counter */
	chip.busy = chip.write_busy
	chip.write_busy_cnt += chip.write_busy
	chip.write_busy =
		cbool.ToInt[uint8]((chip.write_busy != 0 && chip.write_busy_cnt>>5 == 0) || chip.write_d_en != 0)
	chip.write_busy_cnt &= 0x1f
}

func OPN2_DoRegWrite(chip *YM3438) {
	var slot uint32 = chip.cycles % 12
	var address uint32
	var channel uint32 = chip.channel

	/* Update registers */
	if chip.write_fm_data != 0 {
		/* Slot */
		if op_offset[slot] == (uint32(chip.address) & 0x107) {
			if chip.address&0x08 != 0 {
				/* OP2, OP4 */
				slot += 12
			}
			address = uint32(chip.address & 0xf0)
			switch address {
			case 0x30: /* DT, MULTI */
				chip.multi[slot] = chip.data & 0x0f
				if chip.multi[slot] == 0 {
					chip.multi[slot] = 1
				} else {
					chip.multi[slot] <<= 1
				}
				chip.dt[slot] = (chip.data >> 4) & 0x07
			case 0x40: /* TL */
				chip.tl[slot] = chip.data & 0x7f
			case 0x50: /* KS, AR */
				chip.ar[slot] = chip.data & 0x1f
				chip.ks[slot] = (chip.data >> 6) & 0x03
			case 0x60: /* AM, DR */
				chip.dr[slot] = chip.data & 0x1f
				chip.am[slot] = (chip.data >> 7) & 0x01
			case 0x70: /* SR */
				chip.sr[slot] = chip.data & 0x1f
			case 0x80: /* SL, RR */
				chip.rr[slot] = chip.data & 0x0f
				chip.sl[slot] = (chip.data >> 4) & 0x0f
				chip.sl[slot] |= (chip.sl[slot] + 1) & 0x10
			case 0x90: /* SSG-EG */
				chip.ssg_eg[slot] = chip.data & 0x0f
			}
		}

		/* Channel */
		if ch_offset[channel] == uint32(chip.address)&0x103 {
			address = uint32(chip.address) & 0xfc
			switch address {
			case 0xa0:
				chip.fnum[channel] = uint16(chip.data)&0xff | (uint16(chip.reg_a4)&0x07)<<8
				chip.block[channel] = (chip.reg_a4 >> 3) & 0x07
				chip.kcode[channel] =
					uint8(uint32(chip.block[channel]<<2) | fn_note[chip.fnum[channel]>>7])
			case 0xa4:
				chip.reg_a4 = chip.data & 0xff
			case 0xa8:
				chip.fnum_3ch[channel] =
					uint16(chip.data)&0xff | (uint16(chip.reg_ac)&0x07)<<8
				chip.block_3ch[channel] = (chip.reg_ac >> 3) & 0x07
				chip.kcode_3ch[channel] = uint8(uint32(chip.block_3ch[channel]<<2) |
					fn_note[chip.fnum_3ch[channel]>>7])
			case 0xac:
				chip.reg_ac = chip.data & 0xff
			case 0xb0:
				chip.connect[channel] = chip.data & 0x07
				chip.fb[channel] = (chip.data >> 3) & 0x07
			case 0xb4:
				chip.pms[channel] = chip.data & 0x07
				chip.ams[channel] = (chip.data >> 4) & 0x03
				chip.pan_l[channel] = (chip.data >> 7) & 0x01
				chip.pan_r[channel] = (chip.data >> 6) & 0x01
			}
		}
	}

	if chip.write_a_en != 0 || chip.write_d_en != 0 {
		/* Data */
		if chip.write_a_en != 0 {
			chip.write_fm_data = 0
		}

		if chip.write_fm_address != 0 && chip.write_d_en != 0 {
			chip.write_fm_data = 1
		}

		/* Address */
		if chip.write_a_en != 0 {
			if (chip.write_data & 0xf0) != 0x00 {
				/* FM Write */
				chip.address = chip.write_data
				chip.write_fm_address = 1
			} else {
				/* SSG write */
				chip.write_fm_address = 0
			}
		}

		/* FM Mode */
		/* Data */
		if chip.write_d_en != 0 && (chip.write_data&0x100) == 0 {
			switch chip.write_fm_mode_a {
			case 0x21: /* LSI test 1 */
				for i := range 8 {
					chip.mode_test_21[i] = uint8((chip.write_data >> i) & 0x01)
				}
			case 0x22: /* LFO control */
				if (chip.write_data>>3)&0x01 != 0 {
					chip.lfo_en = 0x7f
				} else {
					chip.lfo_en = 0
				}
				chip.lfo_freq = uint8(chip.write_data & 0x07)
			case 0x24: /* Timer A */
				chip.timer_a_reg &= 0x03
				chip.timer_a_reg |= (chip.write_data & 0xff) << 2
			case 0x25:
				chip.timer_a_reg &= 0x3fc
				chip.timer_a_reg |= chip.write_data & 0x03
			case 0x26: /* Timer B */
				chip.timer_b_reg = chip.write_data & 0xff
			case 0x27: /* CSM, Timer control */
				chip.mode_ch3 = uint8((chip.write_data & 0xc0) >> 6)
				chip.mode_csm = cbool.ToInt[uint8](chip.mode_ch3 == 2)
				chip.timer_a_load = uint8(chip.write_data & 0x01)
				chip.timer_a_enable = uint8((chip.write_data >> 2) & 0x01)
				chip.timer_a_reset = uint8((chip.write_data >> 4) & 0x01)
				chip.timer_b_load = uint8((chip.write_data >> 1) & 0x01)
				chip.timer_b_enable = uint8((chip.write_data >> 3) & 0x01)
				chip.timer_b_reset = uint8((chip.write_data >> 5) & 0x01)
			case 0x28: /* Key on/off */
				for i := range 4 {
					chip.mode_kon_operator[i] = uint8((chip.write_data >> (4 + i)) & 0x01)
				}
				if (chip.write_data & 0x03) == 0x03 {
					/* Invalid address */
					chip.mode_kon_channel = 0xff
				} else {
					chip.mode_kon_channel =
						uint8((chip.write_data & 0x03) + ((chip.write_data>>2)&1)*3)
				}
			case 0x2a: /* DAC data */
				chip.dacdata &= 0x01
				chip.dacdata |= int16((chip.write_data ^ 0x80) << 1)
			case 0x2b: /* DAC enable */
				chip.dacen = uint8(chip.write_data >> 7)
			case 0x2c: /* LSI test 2 */
				for i := range 8 {
					chip.mode_test_2c[i] = uint8((chip.write_data >> i) & 0x01)
				}
				chip.dacdata &= 0x1fe
				chip.dacdata |= int16(chip.mode_test_2c[3])
				chip.eg_custom_timer = cbool.ToInt[uint8](chip.mode_test_2c[7] == 0 && chip.mode_test_2c[6] != 0)
			}
		}

		/* Address */
		if chip.write_a_en != 0 {
			chip.write_fm_mode_a = chip.write_data & 0x1ff
		}
	}

	if chip.write_fm_data != 0 {
		chip.data = uint8(chip.write_data & 0xff)
	}
}

func OPN2_PhaseCalcIncrement(chip *YM3438) {
	var ch uint32 = chip.channel
	var slot uint32 = chip.cycles
	var fnum uint32 = uint32(chip.pg_fnum)
	var fnum_h uint32 = fnum >> 4
	var fm uint32
	var basefreq uint32
	var lfo uint8 = chip.lfo_pm
	var lfo_l uint8 = lfo & 0x0f
	var pms uint8 = chip.pms[ch]
	var dt uint8 = chip.dt[slot]
	var dt_l uint8 = dt & 0x03
	var detune uint8 = 0
	var block, note uint8
	var sum, sum_h, sum_l uint8
	var kcode uint8 = chip.pg_kcode

	fnum <<= 1
	/* Apply LFO */
	if lfo_l&0x08 != 0 {
		lfo_l ^= 0x0f
	}
	fm = (fnum_h >> pg_lfo_sh1[pms][lfo_l]) + (fnum_h >> pg_lfo_sh2[pms][lfo_l])
	if pms > 5 {
		fm <<= pms - 5
	}
	fm >>= 2
	if lfo&0x10 != 0 {
		fnum -= fm
	} else {
		fnum += fm
	}
	fnum &= 0xfff

	basefreq = (fnum << chip.pg_block) >> 2

	/* Apply detune */
	if dt_l != 0 {
		if kcode > 0x1c {
			kcode = 0x1c
		}
		block = kcode >> 2
		note = kcode & 0x03
		sum = block + 9 + (cbool.ToInt[uint8](dt_l == 3) | (dt_l & 0x02))
		sum_h = sum >> 1
		sum_l = sum & 0x01
		detune = uint8(pg_detune[(sum_l<<2)|note] >> (9 - sum_h))
	}
	if dt&0x04 != 0 {
		basefreq -= uint32(detune)
	} else {
		basefreq += uint32(detune)
	}
	basefreq &= 0x1ffff
	chip.pg_inc[slot] = (basefreq * uint32(chip.multi[slot])) >> 1
	chip.pg_inc[slot] &= 0xfffff
}

func OPN2_PhaseGenerate(chip *YM3438) {
	var slot uint32
	/* Mask increment */
	slot = (chip.cycles + 20) % 24
	if chip.pg_reset[slot] != 0 {
		chip.pg_inc[slot] = 0
	}
	/* Phase step */
	slot = (chip.cycles + 19) % 24
	if chip.pg_reset[slot] != 0 || chip.mode_test_21[3] != 0 {
		chip.pg_phase[slot] = 0
	}
	chip.pg_phase[slot] += chip.pg_inc[slot]
	chip.pg_phase[slot] &= 0xfffff
}

func OPN2_EnvelopeSSGEG(chip *YM3438) {
	var slot uint32 = chip.cycles
	var direction uint8 = 0
	chip.eg_ssg_pgrst_latch[slot] = 0
	chip.eg_ssg_repeat_latch[slot] = 0
	chip.eg_ssg_hold_up_latch[slot] = 0
	if chip.ssg_eg[slot]&0x08 != 0 {
		direction = chip.eg_ssg_dir[slot]
		if chip.eg_level[slot]&0x200 != 0 {
			/* Reset */
			if (chip.ssg_eg[slot] & 0x03) == 0x00 {
				chip.eg_ssg_pgrst_latch[slot] = 1
			}
			/* Repeat */
			if (chip.ssg_eg[slot] & 0x01) == 0x00 {
				chip.eg_ssg_repeat_latch[slot] = 1
			}
			/* Inverse */
			if (chip.ssg_eg[slot] & 0x03) == 0x02 {
				direction ^= 1
			}
			if (chip.ssg_eg[slot] & 0x03) == 0x03 {
				direction = 1
			}
		}
		/* Hold up */
		if chip.eg_kon_latch[slot] != 0 && ((chip.ssg_eg[slot]&0x07) == 0x05 ||
			(chip.ssg_eg[slot]&0x07) == 0x03) {
			chip.eg_ssg_hold_up_latch[slot] = 1
		}
		direction &= chip.eg_kon[slot]
	}
	chip.eg_ssg_dir[slot] = direction
	chip.eg_ssg_enable[slot] = (chip.ssg_eg[slot] >> 3) & 0x01
	chip.eg_ssg_inv[slot] =
		(chip.eg_ssg_dir[slot] ^ (((chip.ssg_eg[slot] >> 2) & 0x01) &
			((chip.ssg_eg[slot] >> 3) & 0x01))) &
			chip.eg_kon[slot]
}

func OPN2_EnvelopeADSR(chip *YM3438) {
	var slot uint32 = (chip.cycles + 22) % 24

	var nkon uint8 = chip.eg_kon_latch[slot]
	var okon uint8 = chip.eg_kon[slot]
	var kon_event uint8
	var koff_event uint8
	var eg_off uint8
	var level int16
	var nextlevel int16
	var ssg_level int16
	var nextstate uint8 = chip.eg_state[slot]
	var inc int16

	chip.eg_read[0] = uint32(chip.eg_read_inc)
	chip.eg_read_inc = cbool.ToInt[uint8](chip.eg_inc > 0)

	/* Reset phase generator */
	chip.pg_reset[slot] = cbool.ToInt[uint8]((nkon != 0 && okon == 0) || chip.eg_ssg_pgrst_latch[slot] != 0)

	/* KeyOn/Off */
	kon_event = cbool.ToInt[uint8]((nkon != 0 && okon == 0) || (okon != 0 && chip.eg_ssg_repeat_latch[slot] != 0))
	koff_event = cbool.ToInt[uint8](okon != 0 && nkon == 0)

	level = int16(chip.eg_level[slot])
	ssg_level = level

	if chip.eg_ssg_inv[slot] != 0 {
		/* Inverse */
		ssg_level = 512 - level
		ssg_level &= 0x3ff
	}
	if koff_event != 0 {
		level = ssg_level
	}
	if chip.eg_ssg_enable[slot] != 0 {
		eg_off = uint8(level >> 9)
	} else {
		eg_off = cbool.ToInt[uint8]((level & 0x3f0) == 0x3f0)
	}
	nextlevel = level
	if kon_event != 0 {
		nextstate = eg_num_attack
		/* Instant attack */
		if chip.eg_ratemax != 0 {
			nextlevel = 0
		} else if chip.eg_state[slot] == eg_num_attack && level != 0 &&
			chip.eg_inc != 0 && nkon != 0 {
			inc = (^level << chip.eg_inc) >> 5
		}
	} else {
		switch chip.eg_state[slot] {
		case eg_num_attack:
			if level == 0 {
				nextstate = eg_num_decay
			} else if chip.eg_inc != 0 && chip.eg_ratemax == 0 && nkon != 0 {
				inc = (^level << chip.eg_inc) >> 5
			}
		case eg_num_decay:
			if (level >> 4) == int16(chip.eg_sl[1]<<1) {
				nextstate = eg_num_sustain
			} else if eg_off == 0 && chip.eg_inc != 0 {
				inc = 1 << (chip.eg_inc - 1)
				if chip.eg_ssg_enable[slot] != 0 {
					inc <<= 2
				}
			}
		case eg_num_sustain, eg_num_release:
			if eg_off == 0 && chip.eg_inc != 0 {
				inc = 1 << (chip.eg_inc - 1)
				if chip.eg_ssg_enable[slot] != 0 {
					inc <<= 2
				}
			}
		}
		if nkon == 0 {
			nextstate = eg_num_release
		}
	}
	if chip.eg_kon_csm[slot] != 0 {
		nextlevel |= int16(chip.eg_tl[1]) << 3
	}

	/* Envelope off */
	if kon_event == 0 && chip.eg_ssg_hold_up_latch[slot] == 0 &&
		chip.eg_state[slot] != eg_num_attack && eg_off != 0 {
		nextstate = eg_num_release
		nextlevel = 0x3ff
	}

	nextlevel += inc

	chip.eg_kon[slot] = chip.eg_kon_latch[slot]
	chip.eg_level[slot] = uint16(nextlevel) & 0x3ff
	chip.eg_state[slot] = nextstate
}

func OPN2_EnvelopePrepare(chip *YM3438) {
	var rate uint8
	var sum uint8
	var inc uint8 = 0
	var slot uint32 = chip.cycles
	var rate_sel uint8

	/* Prepare increment */
	rate = (chip.eg_rate << 1) + chip.eg_ksv

	rate = min(rate, 0x3f)

	sum = ((rate >> 2) + chip.eg_shift_lock) & 0x0f
	if chip.eg_rate != 0 && chip.eg_quotient == 2 {
		if rate < 48 {
			switch sum {
			case 12:
				inc = 1
			case 13:
				inc = (rate >> 1) & 0x01
			case 14:
				inc = rate & 0x01
			}
		} else {
			inc = uint8(eg_stephi[rate&0x03][chip.eg_timer_low_lock] + uint32(rate)>>2 - 11)
			inc = min(inc, 4)
		}
	}
	chip.eg_inc = inc
	chip.eg_ratemax = cbool.ToInt[uint8]((rate >> 1) == 0x1f)

	/* Prepare rate & ksv */
	rate_sel = chip.eg_state[slot]
	if (chip.eg_kon[slot] != 0 && chip.eg_ssg_repeat_latch[slot] != 0) ||
		(chip.eg_kon[slot] == 0 && chip.eg_kon_latch[slot] != 0) {
		rate_sel = eg_num_attack
	}
	switch rate_sel {
	case eg_num_attack:
		chip.eg_rate = chip.ar[slot]
	case eg_num_decay:
		chip.eg_rate = chip.dr[slot]
	case eg_num_sustain:
		chip.eg_rate = chip.sr[slot]
	case eg_num_release:
		chip.eg_rate = (chip.rr[slot] << 1) | 0x01
	}
	chip.eg_ksv = chip.pg_kcode >> (chip.ks[slot] ^ 0x03)
	if chip.am[slot] != 0 {
		chip.eg_lfo_am = chip.lfo_am >> eg_am_shift[chip.ams[chip.channel]]
	} else {
		chip.eg_lfo_am = 0
	}
	/* Delay TL & SL value */
	chip.eg_tl[1] = chip.eg_tl[0]
	chip.eg_tl[0] = chip.tl[slot]
	chip.eg_sl[1] = chip.eg_sl[0]
	chip.eg_sl[0] = chip.sl[slot]
}

func OPN2_EnvelopeGenerate(chip *YM3438) {
	var slot uint32 = (chip.cycles + 23) % 24
	var level uint16

	level = chip.eg_level[slot]

	if chip.eg_ssg_inv[slot] != 0 {
		/* Inverse */
		level = 512 - level
	}
	if chip.mode_test_21[5] != 0 {
		level = 0
	}
	level &= 0x3ff

	/* Apply AM LFO */
	level += uint16(chip.eg_lfo_am)

	/* Apply TL */
	if !(chip.mode_csm != 0 && chip.channel == 2+1) {
		level += uint16(chip.eg_tl[0]) << 3
	}
	if level > 0x3ff {
		level = 0x3ff
	}
	chip.eg_out[slot] = level
}

func OPN2_UpdateLFO(chip *YM3438) {
	if (uint32(chip.lfo_quotient) & lfo_cycles[chip.lfo_freq]) ==
		lfo_cycles[chip.lfo_freq] {
		chip.lfo_quotient = 0
		chip.lfo_cnt++
	} else {
		chip.lfo_quotient += chip.lfo_inc
	}
	chip.lfo_cnt &= chip.lfo_en
}

func OPN2_FMPrepare(chip *YM3438) {
	var slot uint32 = (chip.cycles + 6) % 24
	var channel uint32 = chip.channel
	var mod, mod1, mod2 int16
	var op uint32 = slot / 6
	var connect uint8 = chip.connect[channel]
	var prevslot uint32 = (chip.cycles + 18) % 24

	/* Calculate modulation */
	mod1 = 0
	mod2 = 0

	if fm_algorithm[op][0][connect] != 0 {
		mod2 |= chip.fm_op1[channel][0]
	}
	if fm_algorithm[op][1][connect] != 0 {
		mod1 |= chip.fm_op1[channel][1]
	}
	if fm_algorithm[op][2][connect] != 0 {
		mod1 |= chip.fm_op2[channel]
	}
	if fm_algorithm[op][3][connect] != 0 {
		mod2 |= chip.fm_out[prevslot]
	}
	if fm_algorithm[op][4][connect] != 0 {
		mod1 |= chip.fm_out[prevslot]
	}
	mod = mod1 + mod2
	if op == 0 {
		/* Feedback */
		mod = mod >> (10 - chip.fb[channel])
		if chip.fb[channel] == 0 {
			mod = 0
		}
	} else {
		mod >>= 1
	}
	chip.fm_mod[slot] = uint16(mod)

	slot = (chip.cycles + 18) % 24
	/* OP1 */
	if slot/6 == 0 {
		chip.fm_op1[channel][1] = chip.fm_op1[channel][0]
		chip.fm_op1[channel][0] = chip.fm_out[slot]
	}
	/* OP2 */
	if slot/6 == 2 {
		chip.fm_op2[channel] = chip.fm_out[slot]
	}
}

func OPN2_ChGenerate(chip *YM3438) {
	var slot uint32 = (chip.cycles + 18) % 24
	var channel uint32 = chip.channel
	var op uint32 = slot / 6
	var test_dac uint32 = uint32(chip.mode_test_2c[5])
	var acc int16 = chip.ch_acc[channel]
	var add int16 = int16(test_dac)
	var sum int16 = 0
	if op == 0 && test_dac == 0 {
		acc = 0
	}
	if fm_algorithm[op][5][chip.connect[channel]] != 0 && test_dac == 0 {
		add += chip.fm_out[slot] >> 5
	}
	sum = acc + add
	/* Clamp */
	if sum > 255 {
		sum = 255
	} else if sum < -256 {
		sum = -256
	}

	if op == 0 || test_dac != 0 {
		chip.ch_out[channel] = chip.ch_acc[channel]
	}
	chip.ch_acc[channel] = sum
}

func OPN2_ChOutput(chip *YM3438) {
	var cycles uint32 = chip.cycles
	var slot uint32 = chip.cycles
	var channel uint32 = chip.channel
	var test_dac uint32 = uint32(chip.mode_test_2c[5])
	var out int16
	var sign int16
	var out_en uint32
	chip.ch_read = chip.ch_lock
	if slot < 12 {
		/* Ch 4,5,6 */
		channel++
	}
	if (cycles & 3) == 0 {
		if test_dac == 0 {
			/* Lock value */
			chip.ch_lock = chip.ch_out[channel]
		}
		chip.ch_lock_l = chip.pan_l[channel]
		chip.ch_lock_r = chip.pan_r[channel]
	}
	/* Ch 6 */
	if ((cycles>>2) == 1 && chip.dacen != 0) || test_dac != 0 {
		out = (int16)(chip.dacdata)
		out = SIGN_EXTEND(8, out)
	} else {
		out = chip.ch_lock
	}
	chip.mol = 0
	chip.mor = 0

	if chip_type&ModeYM2612 != 0 {
		out_en = cbool.ToInt[uint32](((cycles & 3) == 3) || test_dac != 0)
		/* YM2612 DAC emulation(not verified) */
		sign = out >> 8
		if out >= 0 {
			out++
			sign++
		}
		if chip.ch_lock_l != 0 && out_en != 0 {
			chip.mol = out
		} else {
			chip.mol = sign
		}
		if chip.ch_lock_r != 0 && out_en != 0 {
			chip.mor = out
		} else {
			chip.mor = sign
		}
		/* Amplify signal */
		chip.mol *= 3
		chip.mor *= 3
	} else {
		out_en = cbool.ToInt[uint32](((cycles & 3) != 0) || test_dac != 0)
		if chip.ch_lock_l != 0 && out_en != 0 {
			chip.mol = out
		}
		if chip.ch_lock_r != 0 && out_en != 0 {
			chip.mor = out
		}
	}
}

func OPN2_FMGenerate(chip *YM3438) {
	var slot uint32 = (chip.cycles + 19) % 24
	/* Calculate phase */
	var phase uint16 = uint16((uint32(chip.fm_mod[slot]) + (chip.pg_phase[slot] >> 10)) & 0x3ff)
	var quarter uint16
	var level uint16
	var output int16
	if phase&0x100 != 0 {
		quarter = (phase ^ 0xff) & 0xff
	} else {
		quarter = phase & 0xff
	}
	level = logsinrom[quarter]
	/* Apply envelope */
	level += chip.eg_out[slot] << 2
	/* Transform */
	if level > 0x1fff {
		level = 0x1fff
	}
	output = int16(((exprom[(level&0xff)^0xff] | 0x400) << 2) >> (level >> 8))
	if phase&0x200 != 0 {
		output = ((^output) ^ (int16(chip.mode_test_21[4]) << 13)) + 1
	} else {
		output = output ^ (int16(chip.mode_test_21[4]) << 13)
	}
	output <<= 2
	output >>= 2
	chip.fm_out[slot] = output
}

func OPN2_DoTimerA(chip *YM3438) {
	var time uint16
	var load uint8
	load = chip.timer_a_overflow
	if chip.cycles == 2 {
		/* Lock load value */
		load |= cbool.ToInt[uint8](chip.timer_a_load_lock == 0 && chip.timer_a_load != 0)
		chip.timer_a_load_lock = chip.timer_a_load
		if chip.mode_csm != 0 {
			/* CSM KeyOn */
			chip.mode_kon_csm = load
		} else {
			chip.mode_kon_csm = 0
		}
	}
	/* Load counter */
	if chip.timer_a_load_latch != 0 {
		time = chip.timer_a_reg
	} else {
		time = chip.timer_a_cnt
	}
	chip.timer_a_load_latch = load
	/* Increase counter */
	if (chip.cycles == 1 && chip.timer_a_load_lock != 0) || chip.mode_test_21[2] != 0 {
		time++
	}
	/* Set overflow flag */
	if chip.timer_a_reset != 0 {
		chip.timer_a_reset = 0
		chip.timer_a_overflow_flag = 0
	} else {
		chip.timer_a_overflow_flag |= chip.timer_a_overflow & chip.timer_a_enable
	}
	chip.timer_a_overflow = uint8(time >> 10)
	chip.timer_a_cnt = time & 0x3ff
}

func OPN2_DoTimerB(chip *YM3438) {
	var time uint16
	var load uint8
	load = chip.timer_b_overflow
	if chip.cycles == 2 {
		/* Lock load value */
		load |= cbool.ToInt[uint8](chip.timer_b_load_lock == 0 && chip.timer_b_load != 0)
		chip.timer_b_load_lock = chip.timer_b_load
	}
	/* Load counter */
	if chip.timer_b_load_latch != 0 {
		time = chip.timer_b_reg
	} else {
		time = chip.timer_b_cnt
	}
	chip.timer_b_load_latch = load
	/* Increase counter */
	if chip.cycles == 1 {
		chip.timer_b_subcnt++
	}
	if (chip.timer_b_subcnt == 0x10 && chip.timer_b_load_lock != 0) ||
		chip.mode_test_21[2] != 0 {
		time++
	}
	chip.timer_b_subcnt &= 0x0f
	/* Set overflow flag */
	if chip.timer_b_reset != 0 {
		chip.timer_b_reset = 0
		chip.timer_b_overflow_flag = 0
	} else {
		chip.timer_b_overflow_flag |= chip.timer_b_overflow & chip.timer_b_enable
	}
	chip.timer_b_overflow = uint8(time >> 8)
	chip.timer_b_cnt = time & 0xff
}

func OPN2_KeyOn(chip *YM3438) {
	var slot uint32 = chip.cycles
	var ch uint32 = chip.channel
	/* Key On */
	chip.eg_kon_latch[slot] = chip.mode_kon[slot]
	chip.eg_kon_csm[slot] = 0
	if chip.channel == 2 && chip.mode_kon_csm != 0 {
		/* CSM Key On */
		chip.eg_kon_latch[slot] = 1
		chip.eg_kon_csm[slot] = 1
	}
	if chip.cycles == uint32(chip.mode_kon_channel) {
		/* OP1 */
		chip.mode_kon[ch] = chip.mode_kon_operator[0]
		/* OP2 */
		chip.mode_kon[ch+12] = chip.mode_kon_operator[1]
		/* OP3 */
		chip.mode_kon[ch+6] = chip.mode_kon_operator[2]
		/* OP4 */
		chip.mode_kon[ch+18] = chip.mode_kon_operator[3]
	}
}

func OPN2_Reset(chip *YM3438, rate uint32, clock uint32) {
	*chip = YM3438{}
	for i := range 24 {
		chip.eg_out[i] = 0x3ff
		chip.eg_level[i] = 0x3ff
		chip.eg_state[i] = eg_num_release
		chip.multi[i] = 1
	}
	for i := range 6 {
		chip.pan_l[i] = 1
		chip.pan_r[i] = 1
	}

	chip.rateratio = int32(((uint64(144 * rate)) << RSM_FRAC) / uint64(clock))
}

func OPN2_SetChipType(typ uint32) { chip_type = typ }

func OPN2_Clock(chip *YM3438, buffer []int32) {
	var slot uint32 = chip.cycles
	chip.lfo_inc = chip.mode_test_21[1]
	chip.pg_read >>= 1
	chip.eg_read[1] >>= 1
	chip.eg_cycle++
	/* Lock envelope generator timer value */
	if chip.cycles == 1 && chip.eg_quotient == 2 {
		if chip.eg_cycle_stop != 0 {
			chip.eg_shift_lock = 0
		} else {
			chip.eg_shift_lock = chip.eg_shift + 1
		}
		chip.eg_timer_low_lock = uint8(chip.eg_timer & 0x03)
	}
	/* Cycle specific functions */
	switch chip.cycles {
	case 0:
		chip.lfo_pm = chip.lfo_cnt >> 2
		if chip.lfo_cnt&0x40 != 0 {
			chip.lfo_am = chip.lfo_cnt & 0x3f
		} else {
			chip.lfo_am = chip.lfo_cnt ^ 0x3f
		}
		chip.lfo_am <<= 1
	case 1:
		chip.eg_quotient++
		chip.eg_quotient %= 3
		chip.eg_cycle = 0
		chip.eg_cycle_stop = 1
		chip.eg_shift = 0
		chip.eg_timer_inc |= uint8(chip.eg_quotient >> 1)
		chip.eg_timer = chip.eg_timer + uint16(chip.eg_timer_inc)
		chip.eg_timer_inc = uint8(chip.eg_timer >> 12)
		chip.eg_timer &= 0xfff
	case 2:
		chip.pg_read = chip.pg_phase[21] & 0x3ff
		chip.eg_read[1] = uint32(chip.eg_out[0])
	case 13:
		chip.eg_cycle = 0
		chip.eg_cycle_stop = 1
		chip.eg_shift = 0
		chip.eg_timer = chip.eg_timer + uint16(chip.eg_timer_inc)
		chip.eg_timer_inc = uint8(chip.eg_timer >> 12)
		chip.eg_timer &= 0xfff
	case 23:
		chip.lfo_inc |= 1
	}
	chip.eg_timer &= ^uint16(chip.mode_test_21[5] << chip.eg_cycle)
	if ((chip.eg_timer>>chip.eg_cycle)|
		uint16((chip.pin_test_in&chip.eg_custom_timer)))&
		uint16(chip.eg_cycle_stop) != 0 {
		chip.eg_shift = chip.eg_cycle
		chip.eg_cycle_stop = 0
	}

	OPN2_DoIO(chip)

	OPN2_DoTimerA(chip)
	OPN2_DoTimerB(chip)
	OPN2_KeyOn(chip)

	OPN2_ChOutput(chip)
	OPN2_ChGenerate(chip)

	OPN2_FMPrepare(chip)
	OPN2_FMGenerate(chip)

	OPN2_PhaseGenerate(chip)
	OPN2_PhaseCalcIncrement(chip)

	OPN2_EnvelopeADSR(chip)
	OPN2_EnvelopeGenerate(chip)
	OPN2_EnvelopeSSGEG(chip)
	OPN2_EnvelopePrepare(chip)

	/* Prepare fnum & block */
	if chip.mode_ch3 != 0 {
		/* Channel 3 special mode */
		switch slot {
		case 1: /* OP1 */
			chip.pg_fnum = chip.fnum_3ch[1]
			chip.pg_block = chip.block_3ch[1]
			chip.pg_kcode = chip.kcode_3ch[1]
		case 7: /* OP3 */
			chip.pg_fnum = chip.fnum_3ch[0]
			chip.pg_block = chip.block_3ch[0]
			chip.pg_kcode = chip.kcode_3ch[0]
		case 13: /* OP2 */
			chip.pg_fnum = chip.fnum_3ch[2]
			chip.pg_block = chip.block_3ch[2]
			chip.pg_kcode = chip.kcode_3ch[2]
		default: /* OP4 */
			chip.pg_fnum = chip.fnum[(chip.channel+1)%6]
			chip.pg_block = chip.block[(chip.channel+1)%6]
			chip.pg_kcode = chip.kcode[(chip.channel+1)%6]
		}
	} else {
		chip.pg_fnum = chip.fnum[(chip.channel+1)%6]
		chip.pg_block = chip.block[(chip.channel+1)%6]
		chip.pg_kcode = chip.kcode[(chip.channel+1)%6]
	}

	OPN2_UpdateLFO(chip)
	OPN2_DoRegWrite(chip)
	chip.cycles = (chip.cycles + 1) % 24
	chip.channel = chip.cycles % 6

	buffer[0] = int32(chip.mol)
	buffer[1] = int32(chip.mor)

	if chip.status_time != 0 {
		chip.status_time--
	}
}

func OPN2_Write(chip *YM3438, port uint32, data uint8) {
	port &= 3
	chip.write_data = uint16(((port << 7) & 0x100) | uint32(data))
	if port&1 != 0 {
		/* Data */
		chip.write_d |= 1
	} else {
		/* Address */
		chip.write_a |= 1
	}
}

func OPN2_SetTestPin(chip *YM3438, value uint32) {
	chip.pin_test_in = uint8(value & 1)
}

func OPN2_ReadTestPin(chip *YM3438) uint32 {
	if chip.mode_test_2c[7] == 0 {
		return 0
	}

	return cbool.ToInt[uint32](chip.cycles == 23)
}

func OPN2_ReadIRQPin(chip *YM3438) uint32 {
	return uint32(chip.timer_a_overflow_flag | chip.timer_b_overflow_flag)
}

func OPN2_Read(chip *YM3438, port uint32) uint8 {
	if (port&3) == 0 || (chip_type&ModeReadmode != 0) {
		if chip.mode_test_21[6] != 0 {
			/* Read test data */
			var slot uint32 = (chip.cycles + 18) % 24
			var testdata uint16 = uint16(((chip.pg_read & 0x01) << 15) |
				((chip.eg_read[chip.mode_test_21[0]] & 0x01) << 14))
			if chip.mode_test_2c[4] != 0 {
				testdata |= uint16(chip.ch_read & 0x1ff)
			} else {
				testdata |= uint16(chip.fm_out[slot] & 0x3fff)
			}
			if chip.mode_test_21[7] != 0 {
				chip.status = uint8(testdata & 0xff)
			} else {
				chip.status = uint8(testdata >> 8)
			}
		} else {
			chip.status = (chip.busy << 7) | (chip.timer_b_overflow_flag << 1) |
				chip.timer_a_overflow_flag
		}
		if chip_type&ModeYM2612 != 0 {
			chip.status_time = 300000
		} else {
			chip.status_time = 40000000
		}
	}
	if chip.status_time != 0 {
		return chip.status
	}
	return 0
}

func OPN2_WriteBuffered(chip *YM3438, port uint32, data uint8) {
	var time1, time2 uint64
	var buffer [2]int32
	var skip uint64

	if chip.writebuf[chip.writebuf_last].port&0x04 != 0 {
		OPN2_Write(chip, uint32(chip.writebuf[chip.writebuf_last].port&0x03),
			chip.writebuf[chip.writebuf_last].data)

		chip.writebuf_cur = (chip.writebuf_last + 1) % OPN_WRITEBUF_SIZE
		skip = chip.writebuf[chip.writebuf_last].time - chip.writebuf_samplecnt
		chip.writebuf_samplecnt = chip.writebuf[chip.writebuf_last].time
		for ; skip >= 0; skip-- {
			OPN2_Clock(chip, buffer[:])
		}
	}

	chip.writebuf[chip.writebuf_last].port = uint8((port & 0x03) | 0x04)
	chip.writebuf[chip.writebuf_last].data = data
	time1 = chip.writebuf_lasttime + OPN_WRITEBUF_DELAY
	time2 = chip.writebuf_samplecnt

	if time1 < time2 {
		time1 = time2
	}

	chip.writebuf[chip.writebuf_last].time = time1
	chip.writebuf_lasttime = time1
	chip.writebuf_last = (chip.writebuf_last + 1) % OPN_WRITEBUF_SIZE
}

var (
	use_filter = 1 // FIXME(evgenii.omelchenko): should be part of chip
)

func OPN2_GenerateResampled(chip *YM3438, buf []int32) {
	var buffer [2]int32
	var mute uint32

	for chip.samplecnt >= chip.rateratio {
		chip.oldsamples[0] = chip.samples[0]
		chip.oldsamples[1] = chip.samples[1]
		chip.samples[0] = 0
		chip.samples[1] = 0
		for range 24 {
			switch chip.cycles >> 2 {
			case 0: // Ch 2
				mute = chip.mute[1]
			case 1: // Ch 6, DAC
				mute = chip.mute[5+chip.dacen]
			case 2: // Ch 4
				mute = chip.mute[3]
			case 3: // Ch 1
				mute = chip.mute[0]
			case 4: // Ch 5
				mute = chip.mute[4]
			case 5: // Ch 3
				mute = chip.mute[2]
			default:
				mute = 0
			}
			OPN2_Clock(chip, buffer[:])
			if mute == 0 {
				chip.samples[0] += buffer[0]
				chip.samples[1] += buffer[1]
			}

			for chip.writebuf[chip.writebuf_cur].time <=
				chip.writebuf_samplecnt {
				if chip.writebuf[chip.writebuf_cur].port&0x04 == 0 {
					break
				}
				chip.writebuf[chip.writebuf_cur].port &= 0x03
				OPN2_Write(chip, uint32(chip.writebuf[chip.writebuf_cur].port),
					chip.writebuf[chip.writebuf_cur].data)
				chip.writebuf_cur = (chip.writebuf_cur + 1) % OPN_WRITEBUF_SIZE
			}
			chip.writebuf_samplecnt++
		}

		if use_filter == 0 {
			chip.samples[0] *= OUTPUT_FACTOR
			chip.samples[1] *= OUTPUT_FACTOR
		} else {
			chip.samples[0] = int32(float64(chip.oldsamples[0]) +
				FILTER_CUTOFF_I*float64(chip.samples[0]*OUTPUT_FACTOR_F-
					chip.oldsamples[0]))
			chip.samples[1] = int32(float64(chip.oldsamples[1]) +
				FILTER_CUTOFF_I*float64(chip.samples[1]*OUTPUT_FACTOR_F-
					chip.oldsamples[1]))
		}

		chip.samplecnt -= chip.rateratio
	}

	buf[0] = ((chip.oldsamples[0]*(chip.rateratio-chip.samplecnt) +
		chip.samples[0]*chip.samplecnt) /
		chip.rateratio)
	buf[1] = ((chip.oldsamples[1]*(chip.rateratio-chip.samplecnt) +
		chip.samples[1]*chip.samplecnt) /
		chip.rateratio)
	chip.samplecnt += 1 << RSM_FRAC
}

func OPN2_GenerateStream(chip *YM3438, sndptr [][]int32, numsamples uint32) {
	var smpl, smpr []int32
	var buffer [2]int32

	smpl = sndptr[0]
	smpr = sndptr[1]

	for i := range numsamples {
		OPN2_GenerateResampled(chip, buffer[:])

		smpl[i] = buffer[0]
		smpr[i] = buffer[1]
	}
}
