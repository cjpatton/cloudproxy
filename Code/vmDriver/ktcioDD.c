/*
 * File ktcioDD.c - 
 *	Virtualized trusted service device driver. 
 *	This is heavily based on the tcioDD.c driver code.
 *
 * This file and derived words are subject to the terms and conditions
 * set forth in the file LICENSE in this directory.
 */  

#include <linux/module.h>
#include <linux/moduleparam.h>
#include <linux/init.h>

#include <linux/kernel.h> /* printk() */
#include <linux/fs.h>     /* everything... */
#include <linux/errno.h>  /* error codes */
#include <linux/types.h>  /* size_t */
#include <linux/vmalloc.h>
#include <linux/blkdev.h>
#include <linux/hdreg.h>
#include <linux/slab.h>
#include <linux/proc_fs.h>
#include <linux/fcntl.h>
#include <linux/seq_file.h>
#include <linux/cdev.h>
#include <asm/uaccess.h>
#include "ktcioDD.h"
#include <asm/vmdd.h>
//#include <asm/kvm_para.h>

extern ssize_t  ktciodd_read(struct file *filp, char __user *buf, size_t count,
                            loff_t *f_pos);
extern ssize_t  ktciodd_write(struct file *filp, const char __user *buf, size_t count,
                             loff_t *f_pos);
//extern int ktciodd_ioctl(struct inode *inode, struct file *filp, unsigned cmd, unsigned long arg); 

//extern int ktciodd_init(void);
//extern int ktciodd_exit(void);
extern int ktciodd_open(void);
extern int ktciodd_close(void);

int ktciodd_major = KTCIODD_MAJOR;
int ktciodd_minor = 0;
int ktciodd_nr_devs = KTCIODD_NR_DEVS;

module_param(ktciodd_major, int, S_IRUGO);
module_param(ktciodd_minor, int, S_IRUGO);
module_param(ktciodd_nr_devs, int, S_IRUGO);

struct ktciodd_dev*  ktciodd_devices;

struct file_operations ktciodd_fops= {
    .owner=    THIS_MODULE,
    .read=     ktciodd_read,
    .write=    ktciodd_write,
    .open=     ktciodd_open,
    .release=  ktciodd_close,
};

static struct class*    kclass= NULL;
static struct device*   kdevice= NULL;

// Set up the char_dev structure for this device.
void ktciodd_setup_cdev(struct ktciodd_dev *dev, int index)
{
    int err, devno;
   
#ifdef TESTDEVICE
    printk(KERN_DEBUG "tcioDD: setup cdev started\n");
#endif
    devno= MKDEV(ktciodd_major, ktciodd_minor+index);
    cdev_init(&dev->kdev, &ktciodd_fops);
    dev->kdev.owner= THIS_MODULE;
    dev->kdev.ops= &ktciodd_fops;
    err= cdev_add (&dev->kdev, devno, 1);
    if(err)
        printk(KERN_NOTICE "Error %d adding ktciodd %d", err, index);
#ifdef TESTDEVICE
    printk(KERN_DEBUG "ktcioDD: setup cdev complete, devno is %08x\n", devno);
#endif

}

int ktciodd_close(void) {
    int         i;
    dev_t       devno= MKDEV(ktciodd_major, ktciodd_minor);
	int result;

#ifdef TESTDEVICE
    printk(KERN_DEBUG "ktcioDD: close started\n");
#endif
    // Get rid of dev entries
    if(ktciodd_devices) {
        for(i= 0; i<ktciodd_nr_devs; i++) {
            cdev_del(&ktciodd_devices[i].kdev);
        }
        kfree(ktciodd_devices);
    }

    if(kclass!=NULL && kdevice!=NULL) {
        device_destroy(kclass, MKDEV(ktciodd_major,0));
        kdevice= NULL;
    }
    if(kclass!=NULL) {
        class_destroy(kclass);
        kclass= NULL;
    }

    // cleanup_module isn't called if registering failed
    unregister_chrdev_region(devno, ktciodd_nr_devs);
#ifdef TESTDEVICE
    printk(KERN_DEBUG "ktcioDD: close complete\n");
#endif
	result = tc_hypercall0(int, KVM_HYPERCALL_DISCONNECT_FROM_TCSERVICE);
//	res = kvm_hypercall0(KVM_HYPERCALL_DISCONNECT_FROM_TCSERVICE);
	return result;
} //end ktciodd_close

int ktciodd_open(void) {
    int     result, i;
    dev_t   dev= 0;

    if(ktciodd_major) {
        // static registration
        dev= MKDEV(ktciodd_major, ktciodd_minor);
        result= register_chrdev_region(dev, ktciodd_nr_devs, "ktciodd");
    } 
    else {
        // dynamic registration
        result= alloc_chrdev_region(&dev, ktciodd_minor, ktciodd_nr_devs, "ktciodd");
        ktciodd_major= MAJOR(dev);
    }
    if(result<0) {
        printk(KERN_WARNING "ktciodd: can't get major %d\n", ktciodd_major);
        return result;
    }

    kclass= class_create(THIS_MODULE, "ktcioDD");
    if(kclass==NULL)
        goto fail;
    kdevice= device_create(kclass, NULL, MKDEV(ktciodd_major,0), NULL, "ktcioDD0");
    if(kdevice==NULL)
        goto fail;

    ktciodd_devices= kmalloc(ktciodd_nr_devs*sizeof(struct ktciodd_dev), GFP_KERNEL);
    if(!ktciodd_devices) {
        result= -ENOMEM;
        goto fail;
    }
    memset(ktciodd_devices, 0, ktciodd_nr_devs*sizeof(struct ktciodd_dev));

    // Initialize each device. 
    for(i= 0; i<ktciodd_nr_devs; i++) {
        ktciodd_setup_cdev(&ktciodd_devices[i], i);
    }
#ifdef TESTDEVICE
    printk(KERN_DEBUG "ktcioDD: ktciodd_init complete\n");
#endif
	result = tc_hypercall0(int, KVM_HYPERCALL_CONNECT_TO_TCSERVICE);
//	result = kvm_hypercall0(KVM_HYPERCALL_CONNECT_TO_TCSERVICE);
	return result;
//    return 0;

fail:
    ktciodd_close();
    return result;
} //end ktciodd_open

ssize_t ktciodd_read(struct file *filp, char __user *buf, size_t count,
                    loff_t *f_pos) {

	ssize_t result = 0;
	result = tc_hypercall4(ssize_t, KVM_HYPERCALL_READ_FROM_TCSERVICE, filp, buf, count, f_pos);
//	result = kvm_hypercall4(KVM_HYPERCALL_READ_FROM_TCSERVICE, (unsigned long) filp, (unsigned long) buf, (unsigned long) count, (unsigned long) f_pos);
	//	return tciodd_read(filp, buf, count, f_pos);
/*
 * REK: there are two ways to communicate to the tcService:
 * 	a) Use a port via serial interface between the guest and host
 *	b) Do a hypercall to invoke host, and handle the hypercall to 
 *	pass the message to tcService
 */
	return result;
//	return 0;
} //end ktciodd_read

ssize_t ktciodd_write(struct file *filp, const char __user *buf, size_t count,
                     loff_t *f_pos) {

	ssize_t result = 0;
/*
 * REK: there are two ways to communicate to the tcService:
 * 	a) Use a port via serial interface between the guest and host
 *	b) Do a hypercall to invoke host, and handle the hypercall to 
 *	pass the message to tcService
 */
	result = tc_hypercall4(ssize_t, KVM_HYPERCALL_WRITE_TO_TCSERVICE, filp, buf, count, f_pos);
//	result = kvm_hypercall4(KVM_HYPERCALL_WRITE_TO_TCSERVICE, 
//				(unsigned long) filp, (unsigned long) buf, 
//				(unsigned long) count, (unsigned long) f_pos);
//		return tciodd_write(filp, buf, count, f_pos);
	return result;
//	return 0;
}

#ifdef LINUXLICENSED
MODULE_LICENSE("GPL");
module_init(ktciodd_open);
module_exit(ktciodd_close);
#endif
